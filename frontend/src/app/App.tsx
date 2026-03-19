import { FormEvent, ReactNode, useEffect, useMemo, useState } from 'react';
import {
  ArrowRight,
  CalendarClock,
  CheckCircle2,
  Database,
  Download,
  FileSpreadsheet,
  LogOut,
  Minus,
  MonitorSmartphone,
  Moon,
  Plus,
  Search,
  Sparkles,
  Sun,
  Upload,
  Wallet,
} from 'lucide-react';
import {
  APIError,
  createKeywordSeed,
  generateRecommendation,
  getAnonymousSession,
  getPublicBootstrap,
  getJDScheduleConfig,
  getKeywordSeed,
  importKeywordSeeds,
  listKeywordSeeds,
  loginAdmin,
  logoutAdmin,
  setKeywordSeedEnabled,
  updateJDScheduleConfig,
  updateKeywordSeed,
  verifyChallenge,
} from './lib/api';
import type {
  CatalogRecommendationResponse,
  CollectorScheduleUpsertRequest,
  GenerateBuildRequest,
  KeywordSeed,
  KeywordSeedImportResponse,
  KeywordSeedUpsertRequest,
  PublicBootstrapResponse,
} from './lib/types';

type PublicFormState = {
  budget: string;
  useCase: string;
  buildMode: string;
  cpuBrand: string;
  gpuBrand: string;
  specialRequirements: string;
  notes: string;
};

type KeywordFormState = {
  category: string;
  keyword: string;
  canonicalModel: string;
  brand: string;
  aliases: string;
  priority: string;
  enabled: boolean;
  notes: string;
};

const defaultPublicForm: PublicFormState = {
  budget: '6000',
  useCase: 'gaming',
  buildMode: 'mixed',
  cpuBrand: '',
  gpuBrand: '',
  specialRequirements: '',
  notes: '',
};

const defaultKeywordForm: KeywordFormState = {
  category: 'cpu',
  keyword: '',
  canonicalModel: '',
  brand: '',
  aliases: '',
  priority: '100',
  enabled: true,
  notes: '',
};

const categoryOptions = ['cpu', 'gpu', 'motherboard', 'ram', 'ssd', 'psu', 'case', 'cooler'];
const publicUseCaseOptions = [
  { value: 'gaming', label: '游戏' },
  { value: 'office', label: '办公' },
  { value: 'design', label: '设计 / 剪辑' },
];
const themeStorageKey = 'givezj8-theme';

declare global {
  interface Window {
    turnstile?: {
      render: (container: HTMLElement, options: { sitekey: string; callback: (token: string) => void }) => string;
      remove: (widgetID: string) => void;
    };
  }
}

export default function App() {
  const pathname = window.location.pathname;
  const editMatch = pathname.match(/^\/admin\/keywords\/([^/]+)\/edit$/);

  if (pathname === '/admin/login') {
    return <AdminLoginPage />;
  }
  if (pathname === '/admin') {
    return <AdminHomePage />;
  }
  if (pathname === '/admin/keywords') {
    return <KeywordListPage />;
  }
  if (pathname === '/admin/keywords/new') {
    return <KeywordFormPage />;
  }
  if (pathname === '/admin/keywords/import') {
    return <KeywordImportPage />;
  }
  if (pathname === '/admin/jd-schedule') {
    return <JDSchedulePage />;
  }
  if (editMatch) {
    return <KeywordFormPage editID={decodeURIComponent(editMatch[1])} />;
  }
  return <RecommendationPage />;
}

function RecommendationPage() {
  useDocumentTitle('给我装机吧 · AI 装机推荐');
  const { isLight, toggleTheme } = useThemeMode();
  const [form, setForm] = useState<PublicFormState>(defaultPublicForm);
  const [anonymousID, setAnonymousID] = useState('');
  const [remaining, setRemaining] = useState('-');
  const [cooldownSeconds, setCooldownSeconds] = useState(0);
  const [requestStatus, setRequestStatus] = useState('等待生成');
  const [isLoading, setIsLoading] = useState(false);
  const [isVerifyingChallenge, setIsVerifyingChallenge] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<CatalogRecommendationResponse | null>(null);
  const [bootstrap, setBootstrap] = useState<PublicBootstrapResponse>({});
  const [challengeToken, setChallengeToken] = useState('');
  const [challengeVisible, setChallengeVisible] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getPublicBootstrap()
      .then((payload) => {
        if (!cancelled) {
          setBootstrap(payload);
        }
      })
      .catch(() => undefined);
    getAnonymousSession()
      .then((session) => {
        if (cancelled) {
          return;
        }
        setAnonymousID(session.anonymous_id);
        setRemaining(String(session.remaining_ai_requests ?? '-'));
        setCooldownSeconds(session.cooldown_seconds ?? 0);
        setChallengeVisible(Boolean(session.challenge_required));
      })
      .catch(() => {
        if (!cancelled) {
          setRequestStatus('匿名会话初始化失败');
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  useEffect(() => {
    if (cooldownSeconds <= 0) {
      return;
    }
    const timer = window.setInterval(() => {
      setCooldownSeconds((current) => {
        if (current <= 1) {
          window.clearInterval(timer);
          return 0;
        }
        return current - 1;
      });
    }, 1000);
    return () => {
      window.clearInterval(timer);
    };
  }, [cooldownSeconds]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsLoading(true);
    setError('');
    setRequestStatus('请求中');
    const payload: GenerateBuildRequest = {
      budget: Number(form.budget),
      use_case: form.useCase,
      build_mode: form.buildMode,
      brand_preference: {
        cpu: form.cpuBrand.trim(),
        gpu: form.gpuBrand.trim(),
      },
      special_requirements: form.specialRequirements
        .split(',')
        .map((item) => item.trim())
        .filter(Boolean),
      notes: form.notes.trim(),
    };

    try {
      const data = await generateRecommendation(payload, anonymousID);
      setResult(data);
      setRemaining(String(data.request_status.remaining_ai_requests ?? '-'));
      setCooldownSeconds(data.request_status.cooldown_seconds ?? 0);
      setRequestStatus(data.request_status.cache_hit ? '已命中缓存' : '已生成新结果');
    } catch (err) {
      const message = err instanceof APIError ? err.message : '生成推荐失败';
      const cooldownSeconds = err instanceof APIError ? err.cooldownSeconds : 0;
      if (err instanceof APIError && err.challengeRequired) {
        setChallengeVisible(true);
      }
      setError(message);
      setCooldownSeconds(cooldownSeconds);
      setRequestStatus(message);
    } finally {
      setIsLoading(false);
    }
  }

  async function handleVerifyChallenge() {
    if (!challengeToken.trim()) {
      setError('请先完成安全验证。');
      return;
    }
    setIsVerifyingChallenge(true);
    setError('');
    try {
      await verifyChallenge(anonymousID, challengeToken.trim());
      setChallengeVisible(false);
      setChallengeToken('');
      setRequestStatus('安全验证已通过，可以继续生成。');
    } catch (err) {
      const message = err instanceof APIError ? err.message : '安全验证失败';
      setError(message);
      setRequestStatus(message);
    } finally {
      setIsVerifyingChallenge(false);
    }
  }

  const selectedItems = result?.selection.selected_items ?? [];
  const advice = result?.advice;
  const summaryText = useMemo(() => {
    if (result?.catalog_warnings?.length) {
      return result.catalog_warnings.join('；');
    }
    if (advice?.summary) {
      return advice.summary;
    }
    return '提交预算和用途后，这里会展示总价、配件清单、推荐理由和价格依据。';
  }, [advice?.summary, result?.catalog_warnings]);

  return (
    <div className="app-root app-page-shell relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_rgba(55,197,255,0.12),_transparent_28%),linear-gradient(180deg,_rgba(5,12,22,0.92),_rgba(5,12,22,1))]">
      <div className="pointer-events-none absolute inset-0 opacity-55 [background-image:linear-gradient(rgba(109,184,255,0.12)_1px,transparent_1px),linear-gradient(90deg,rgba(109,184,255,0.12)_1px,transparent_1px)] [background-size:36px_36px]" />
      <div className="pointer-events-none absolute inset-x-0 top-0 h-80 bg-[radial-gradient(circle_at_top,rgba(95,224,255,0.22),transparent_55%)]" />

      <header className="app-header sticky top-0 z-20 border-b border-cyan-400/10 bg-slate-950/70 backdrop-blur-md">
        <div className="container mx-auto flex items-center justify-between gap-4 px-4 py-3">
          <div className="flex min-w-0 items-center gap-3">
            <div className="rounded-lg border border-cyan-400/20 bg-cyan-400/10 p-2 shadow-[0_0_24px_rgba(95,224,255,0.12)]">
              <Sparkles className="size-5 text-white sm:size-6" />
            </div>
            <div className="min-w-0">
              <p className="app-brand-mark text-[10px] uppercase tracking-[0.34em] text-cyan-300/70 sm:text-xs">givezj8.cn</p>
              <h1 className="text-xl font-bold text-slate-50 sm:text-2xl">给我装机吧</h1>
              <p className="text-xs text-slate-400 sm:text-sm">基于当前硬件价格，快速生成装机建议</p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <ThemeToggleButton isLight={isLight} onToggle={toggleTheme} />
            <NavLink href="/admin/login">后台管理</NavLink>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-4 sm:py-6 lg:h-[calc(100vh-81px)] lg:py-4">
        <section className="app-main-panel mx-auto max-w-7xl rounded-[28px] border border-cyan-300/28 bg-slate-950/78 p-4 shadow-[0_24px_80px_rgba(0,0,0,0.28),inset_0_0_0_1px_rgba(130,220,255,0.08)] backdrop-blur sm:p-6 lg:h-full lg:overflow-hidden">
          <div className="grid gap-5 lg:h-full lg:grid-cols-[minmax(0,1.05fr)_minmax(320px,0.95fr)] lg:items-start">
            <div className={scrollPanelClassName}>
              <section>
                <p className="app-kicker text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300/72">无需注册，直接开始</p>
                <h2 className="mt-2 max-w-3xl text-xl leading-[1.08] text-slate-50 sm:text-2xl lg:text-[1.7rem]">
                  输入预算，直接生成装机方案。
                </h2>
              </section>

              <div className="flex flex-wrap gap-2 text-xs text-slate-400">
                <span className="app-chip rounded-full border border-cyan-300/18 bg-cyan-400/8 px-3 py-1.5">基于当前京东硬件价格整理</span>
              </div>

              <section className="app-form-panel rounded-[24px] border border-cyan-300/24 bg-slate-950/82 p-4 shadow-[0_16px_48px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)] sm:p-5">
                <div className="mb-4">
                  <h3 className="text-lg font-semibold text-slate-50 sm:text-xl">先填需求，再出结果</h3>
                  <p className="mt-2 text-sm leading-6 text-slate-400">
                    先用预算和用途确定主方案，再用可选偏好把结果收窄到更适合你的方向。
                  </p>
                </div>

                <form onSubmit={handleSubmit} className="space-y-4 sm:space-y-5">
                  <Field label="预算 (¥)" icon={<Wallet className="size-4" />}>
                    <div className="app-budget-stepper flex items-center rounded-xl border border-cyan-300/18 bg-slate-900/72 focus-within:border-cyan-200/50">
                      <button
                        type="button"
                        className="flex h-12 w-12 items-center justify-center border-r border-cyan-300/14 text-slate-300 transition hover:bg-white/5 hover:text-slate-50"
                        onClick={() =>
                          setForm((prev) => ({
                            ...prev,
                            budget: String(Math.max(1000, Number(prev.budget || '0') - 100)),
                          }))
                        }
                      >
                        <Minus className="size-4" />
                      </button>
                      <input
                        className={`${controlClassName} rounded-none border-0 bg-transparent text-center [appearance:textfield] focus:border-0 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none`}
                        type="number"
                        min="1000"
                        step="100"
                        value={form.budget}
                        onChange={(event) => setForm((prev) => ({ ...prev, budget: event.target.value }))}
                        required
                      />
                      <button
                        type="button"
                        className="flex h-12 w-12 items-center justify-center border-l border-cyan-300/14 text-slate-300 transition hover:bg-white/5 hover:text-slate-50"
                        onClick={() =>
                          setForm((prev) => ({
                            ...prev,
                            budget: String(Math.max(1000, Number(prev.budget || '0') + 100)),
                          }))
                        }
                      >
                        <Plus className="size-4" />
                      </button>
                    </div>
                  </Field>

                  <div className="grid gap-4">
                    <Field label="用途" icon={<MonitorSmartphone className="size-4" />}>
                      <div className="grid gap-2 sm:grid-cols-3">
                        {publicUseCaseOptions.map((option) => {
                          const active = form.useCase === option.value;
                          return (
                            <button
                              key={option.value}
                              type="button"
                              className={`app-choice-button h-12 rounded-xl border px-4 text-sm font-medium transition ${
                                active
                                  ? 'app-choice-button-active border-cyan-200/60 bg-cyan-300/18 text-cyan-100 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.12)]'
                                  : 'border-cyan-300/18 bg-slate-900/72 text-slate-300 hover:border-cyan-200/35 hover:text-slate-100'
                              }`}
                              onClick={() => setForm((prev) => ({ ...prev, useCase: option.value }))}
                            >
                              {option.label}
                            </button>
                          );
                        })}
                      </div>
                    </Field>
                  </div>

                  <div className="rounded-xl border border-cyan-300/24 bg-cyan-400/5 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
                    <p className="text-sm font-medium text-slate-100">补充说明</p>
                    <p className="mt-1 text-sm leading-6 text-slate-400">有额外要求再写，没有可以留空。</p>
                    <textarea
                      className={`${fieldClassName} mt-3 min-h-24 resize-y px-4 py-3`}
                      placeholder="例如：1080p 游戏为主，希望价格稳一点。"
                      value={form.notes}
                      onChange={(event) => setForm((prev) => ({ ...prev, notes: event.target.value }))}
                    />
                  </div>

                  <button
                    type="submit"
                    className="app-submit-button flex h-12 w-full items-center justify-center rounded-xl bg-cyan-300 text-sm font-semibold text-slate-950 transition hover:bg-cyan-200 disabled:cursor-not-allowed disabled:bg-cyan-100"
                    disabled={isLoading || cooldownSeconds > 0 || challengeVisible}
                  >
                    {isLoading
                      ? '生成中...'
                      : challengeVisible
                        ? '请先完成安全验证'
                        : cooldownSeconds > 0
                          ? `冷却中 ${cooldownSeconds} 秒`
                          : '生成配置方案'}
                  </button>

                  <p className="text-xs text-slate-500">
                    {requestStatus} · 剩余次数 {remaining} · 冷却 {cooldownSeconds} 秒
                  </p>
                  {challengeVisible ? (
                    <ChallengePanel
                      provider={bootstrap.challenge_provider}
                      siteKey={bootstrap.challenge_site_key}
                      token={challengeToken}
                      onTokenChange={setChallengeToken}
                      onVerify={handleVerifyChallenge}
                      verifying={isVerifyingChallenge}
                    />
                  ) : null}
                </form>
              </section>
            </div>

              <section className={`${scrollPanelClassName} app-result-region`}>
                <ResultPanel result={result} summaryText={summaryText} selectedItems={selectedItems} error={error} />
              </section>
          </div>
        </section>
      </main>

      <footer className="app-footer border-t border-cyan-400/10 bg-slate-950/70 backdrop-blur-sm lg:hidden">
        <div className="container mx-auto px-4 py-6 text-center text-sm text-slate-500">
          <p>© 2026 给我装机吧 · givezj8.cn · 基于当前硬件价格生成装机建议</p>
        </div>
      </footer>
    </div>
  );
}

function ChallengePanel({
  provider,
  siteKey,
  token,
  onTokenChange,
  onVerify,
  verifying,
}: {
  provider?: string;
  siteKey?: string;
  token: string;
  onTokenChange: (token: string) => void;
  onVerify: () => void;
  verifying: boolean;
}) {
  return (
    <div className="rounded-xl border border-amber-300/30 bg-amber-400/10 p-4 text-sm text-amber-50">
      <p className="font-semibold">当前请求需要先完成安全验证</p>
      <p className="mt-2 leading-6 text-amber-100/80">
        系统检测到当前请求风险较高，验证通过后才能继续触发高成本推荐。
      </p>
      {provider === 'turnstile' && siteKey ? (
        <TurnstileWidget siteKey={siteKey} onToken={onTokenChange} />
      ) : null}
      <input
        className={`${fieldClassName} mt-3 px-4 py-3`}
        placeholder="挑战 token"
        value={token}
        onChange={(event) => onTokenChange(event.target.value)}
      />
      <button
        type="button"
        className="mt-3 inline-flex h-11 items-center justify-center rounded-xl bg-amber-200 px-4 text-sm font-semibold text-slate-950 transition hover:bg-amber-100 disabled:cursor-not-allowed disabled:bg-amber-50"
        onClick={onVerify}
        disabled={verifying || !token.trim()}
      >
        {verifying ? '验证中...' : '验证后继续'}
      </button>
    </div>
  );
}

function TurnstileWidget({ siteKey, onToken }: { siteKey: string; onToken: (token: string) => void }) {
  useEffect(() => {
    let widgetID = '';
    let cancelled = false;

    function mount() {
      const container = document.getElementById('turnstile-container');
      if (!container || !window.turnstile) {
        return;
      }
      container.innerHTML = '';
      widgetID = window.turnstile.render(container, {
        sitekey: siteKey,
        callback: (token) => {
          if (!cancelled) {
            onToken(token);
          }
        },
      });
    }

    if (window.turnstile) {
      mount();
    } else {
      const script = document.createElement('script');
      script.src = 'https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit';
      script.async = true;
      script.defer = true;
      script.onload = mount;
      document.head.appendChild(script);
    }

    return () => {
      cancelled = true;
      if (widgetID && window.turnstile) {
        window.turnstile.remove(widgetID);
      }
    };
  }, [onToken, siteKey]);

  return <div id="turnstile-container" className="mt-3 min-h-16" />;
}

function AdminLoginPage() {
  useDocumentTitle('givezj8.cn · 后台登录');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [status, setStatus] = useState('等待登录');
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setStatus('登录中');
    try {
      await loginAdmin(username.trim(), password);
      window.location.href = '/admin';
    } catch (err) {
      setStatus(err instanceof APIError ? err.message : '登录失败');
      setSubmitting(false);
    }
  }

  return (
    <AdminScaffold
      title="后台登录"
      description="后台词库管理、导入导出和后续采集操作都从这里进入。前台匿名推荐与后台管理已经统一到同一套 React 页面体系。"
    >
      <div className="mx-auto max-w-xl rounded-[28px] border border-cyan-300/24 bg-slate-950/84 p-6 shadow-[0_28px_72px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)]">
        <form className="space-y-5" onSubmit={handleSubmit}>
          <Field label="用户名">
            <input
              className={controlClassName}
              value={username}
              onChange={(event) => setUsername(event.target.value)}
              autoComplete="username"
              required
            />
          </Field>
          <Field label="密码">
            <input
              className={controlClassName}
              type="password"
              value={password}
              onChange={(event) => setPassword(event.target.value)}
              autoComplete="current-password"
              required
            />
          </Field>
          <button
            className="flex h-12 w-full items-center justify-center rounded-xl bg-cyan-300 text-sm font-semibold text-slate-950 transition hover:bg-cyan-200 disabled:cursor-not-allowed disabled:bg-cyan-100"
            type="submit"
            disabled={submitting}
          >
            {submitting ? '登录中...' : '登录后台'}
          </button>
        </form>
        <p className="mt-4 text-sm text-slate-400">{status}</p>
        <div className="mt-6 flex flex-wrap gap-2">
          <NavLink href="/">返回前台推荐页</NavLink>
        </div>
      </div>
    </AdminScaffold>
  );
}

function AdminHomePage() {
  useDocumentTitle('givezj8.cn · 后台管理');
  async function handleLogout() {
    await logoutAdmin();
    window.location.href = '/admin/login';
  }

  return (
    <AdminScaffold
      title="后台管理"
      description="词库维护、导入导出和管理入口都保留原有能力，只是页面层切换成了 React。"
      toolbar={
        <>
          <NavLink href="/">前台推荐页</NavLink>
          <button className={secondaryButtonClassName} type="button" onClick={handleLogout}>
            <LogOut className="size-4" />
            退出登录
          </button>
        </>
      }
    >
      <div className="grid gap-4 md:grid-cols-3">
        <AdminFeatureCard
          href="/admin/keywords"
          icon={<Database className="size-5 text-cyan-300" />}
          title="词库列表"
          text="查看、筛选、启停现有 keyword seeds。"
        />
        <AdminFeatureCard
          href="/admin/keywords/new"
          icon={<Sparkles className="size-5 text-cyan-300" />}
          title="新增词条"
          text="手动补充主流型号和别名。"
        />
        <AdminFeatureCard
          href="/admin/keywords/import"
          icon={<FileSpreadsheet className="size-5 text-cyan-300" />}
          title="Excel 导入"
          text="下载模板、上传工作簿、查看导入结果。"
        />
        <AdminFeatureCard
          href="/admin/jd-schedule"
          icon={<CalendarClock className="size-5 text-cyan-300" />}
          title="JD 定时采集"
          text="配置每日执行时间、请求间隔和单次查询数量。"
        />
      </div>
    </AdminScaffold>
  );
}

function KeywordListPage() {
  useDocumentTitle('givezj8.cn · 后台词库列表');
  const [category, setCategory] = useState('');
  const [brand, setBrand] = useState('');
  const [keyword, setKeyword] = useState('');
  const [enabled, setEnabled] = useState('');
  const [items, setItems] = useState<KeywordSeed[]>([]);
  const [status, setStatus] = useState('读取中');
  const [loading, setLoading] = useState(true);

  async function loadSeeds() {
    setLoading(true);
    const query = new URLSearchParams();
    if (category) query.set('category', category);
    if (brand.trim()) query.set('brand', brand.trim());
    if (keyword.trim()) query.set('keyword', keyword.trim());
    if (enabled) query.set('enabled', enabled);

    try {
      const data = await listKeywordSeeds(query);
      setItems(data.items);
      setStatus(`共 ${data.total} 条`);
    } catch (err) {
      if (err instanceof APIError && err.status === 401) {
        window.location.href = '/admin/login';
        return;
      }
      setItems([]);
      setStatus(err instanceof APIError ? err.message : '读取词库失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    loadSeeds().catch(() => {
      // handled inside loadSeeds
    });
  }, []);

  async function toggleSeed(item: KeywordSeed) {
    try {
      await setKeywordSeedEnabled(item.id, !item.enabled);
      await loadSeeds();
    } catch (err) {
      setStatus(err instanceof APIError ? err.message : '操作失败');
    }
  }

  return (
    <AdminScaffold
      title="词库列表"
      description="筛选、启停、导出和编辑都保持原有接口与行为。"
      toolbar={
        <>
          <NavLink href="/admin">后台首页</NavLink>
          <NavLink href="/admin/keywords/new">新增词条</NavLink>
          <NavLink href="/admin/keywords/import">Excel 导入</NavLink>
          <NavLink href="/admin/api/v1/keyword-seeds/export">导出 Excel</NavLink>
        </>
      }
    >
      <div className="grid gap-3 rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-5 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)] lg:grid-cols-[1fr_1fr_1fr_220px_auto]">
        <input className={controlClassName} placeholder="类别" value={category} onChange={(event) => setCategory(event.target.value)} list="category-options" />
        <input className={controlClassName} placeholder="品牌" value={brand} onChange={(event) => setBrand(event.target.value)} />
        <input className={controlClassName} placeholder="关键词" value={keyword} onChange={(event) => setKeyword(event.target.value)} />
        <select className={controlClassName} value={enabled} onChange={(event) => setEnabled(event.target.value)}>
          <option value="">启用状态</option>
          <option value="true">true</option>
          <option value="false">false</option>
        </select>
        <button className={primaryButtonClassName} type="button" onClick={() => void loadSeeds()}>
          <Search className="size-4" />
          刷新
        </button>
        <datalist id="category-options">
          {categoryOptions.map((value) => (
            <option value={value} key={value} />
          ))}
        </datalist>
      </div>

      <section className="mt-5 rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-5 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)]">
        <div className="flex items-center justify-between gap-4 border-b border-cyan-300/12 pb-4">
          <p className="text-sm text-slate-400">{status}</p>
          {loading ? <StatusPill>读取中</StatusPill> : null}
        </div>
        <div className="mt-4 space-y-4">
          {items.map((item) => (
            <article
              key={item.id}
              className="rounded-2xl border border-cyan-300/18 bg-slate-900/74 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]"
            >
              <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
                <div className="space-y-2">
                  <div className="flex flex-wrap items-center gap-2">
                    <span className="rounded-full border border-cyan-300/18 bg-cyan-400/8 px-3 py-1 text-xs uppercase tracking-[0.18em] text-cyan-200">
                      {item.category}
                    </span>
                    <span className="text-sm text-slate-400">{item.id}</span>
                  </div>
                  <h3 className="text-lg font-semibold text-slate-100">{item.keyword}</h3>
                  <p className="text-sm text-slate-300">标准型号：{item.canonical_model}</p>
                  <p className="text-sm text-slate-400">
                    品牌：{item.brand || '-'} · 优先级：{item.priority} · 状态：{String(item.enabled)}
                  </p>
                  <p className="text-sm text-slate-500">
                    别名：{item.aliases?.length ? item.aliases.join(' / ') : '无'}
                  </p>
                  {item.notes ? <p className="text-sm leading-6 text-slate-400">{item.notes}</p> : null}
                </div>
                <div className="flex flex-wrap gap-2">
                  <NavLink href={`/admin/keywords/${item.id}/edit`}>编辑</NavLink>
                  <button className={secondaryButtonClassName} type="button" onClick={() => void toggleSeed(item)}>
                    {item.enabled ? '停用' : '启用'}
                  </button>
                </div>
              </div>
            </article>
          ))}
          {!items.length && !loading ? (
            <div className="rounded-2xl border border-dashed border-cyan-300/18 bg-slate-900/50 p-6 text-sm text-slate-400">
              当前没有匹配的词条。
            </div>
          ) : null}
        </div>
      </section>
    </AdminScaffold>
  );
}

function KeywordFormPage({ editID }: { editID?: string }) {
  useDocumentTitle(editID ? 'givezj8.cn · 编辑词条' : 'givezj8.cn · 新增词条');
  const [form, setForm] = useState<KeywordFormState>(defaultKeywordForm);
  const [status, setStatus] = useState('等待提交');
  const [loading, setLoading] = useState(Boolean(editID));
  const isEdit = Boolean(editID);

  useEffect(() => {
    if (!editID) {
      return;
    }
    let cancelled = false;
    getKeywordSeed(editID)
      .then((item) => {
        if (cancelled) {
          return;
        }
        setForm({
          category: item.category,
          keyword: item.keyword,
          canonicalModel: item.canonical_model,
          brand: item.brand || '',
          aliases: (item.aliases || []).join(','),
          priority: String(item.priority),
          enabled: item.enabled,
          notes: item.notes || '',
        });
        setStatus(`已加载 ${item.id}`);
      })
      .catch((err) => {
        setStatus(err instanceof APIError ? err.message : '读取词条失败');
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, [editID]);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const payload: KeywordSeedUpsertRequest = {
      category: form.category,
      keyword: form.keyword.trim(),
      canonical_model: form.canonicalModel.trim(),
      brand: form.brand.trim(),
      aliases: form.aliases.split(',').map((item) => item.trim()).filter(Boolean),
      priority: Number(form.priority || 100),
      enabled: form.enabled,
      notes: form.notes.trim(),
    };
    try {
      const saved = isEdit && editID ? await updateKeywordSeed(editID, payload) : await createKeywordSeed(payload);
      setStatus(`已保存 ${saved.id}`);
      if (!isEdit) {
        window.location.href = `/admin/keywords/${saved.id}/edit`;
      }
    } catch (err) {
      setStatus(err instanceof APIError ? err.message : '保存失败');
    }
  }

  return (
    <AdminScaffold
      title={isEdit ? '编辑词条' : '新增词条'}
      description="类别、关键词、标准型号、别名、优先级和启用状态都继续对应原来的后台接口字段。"
      toolbar={
        <>
          <NavLink href="/admin">后台首页</NavLink>
          <NavLink href="/admin/keywords">返回列表</NavLink>
        </>
      }
    >
      <section className="rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-6 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)]">
        {loading ? (
          <p className="text-sm text-slate-400">加载中...</p>
        ) : (
          <form className="space-y-5" onSubmit={handleSubmit}>
            <div className="grid gap-5 md:grid-cols-2">
              <Field label="类别">
                <select className={controlClassName} value={form.category} onChange={(event) => setForm((prev) => ({ ...prev, category: event.target.value }))}>
                  {categoryOptions.map((value) => (
                    <option value={value} key={value}>
                      {value}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="品牌">
                <input className={controlClassName} value={form.brand} onChange={(event) => setForm((prev) => ({ ...prev, brand: event.target.value }))} />
              </Field>
            </div>
            <Field label="关键词">
              <input className={controlClassName} value={form.keyword} onChange={(event) => setForm((prev) => ({ ...prev, keyword: event.target.value }))} required />
            </Field>
            <Field label="标准型号">
              <input
                className={controlClassName}
                value={form.canonicalModel}
                onChange={(event) => setForm((prev) => ({ ...prev, canonicalModel: event.target.value }))}
                required
              />
            </Field>
            <div className="grid gap-5 md:grid-cols-2">
              <Field label="别名">
                <input className={controlClassName} value={form.aliases} onChange={(event) => setForm((prev) => ({ ...prev, aliases: event.target.value }))} />
              </Field>
              <Field label="优先级">
                <input
                  className={controlClassName}
                  type="number"
                  value={form.priority}
                  onChange={(event) => setForm((prev) => ({ ...prev, priority: event.target.value }))}
                />
              </Field>
            </div>
            <div className="grid gap-5 md:grid-cols-2">
              <Field label="启用状态">
                <select
                  className={controlClassName}
                  value={String(form.enabled)}
                  onChange={(event) => setForm((prev) => ({ ...prev, enabled: event.target.value === 'true' }))}
                >
                  <option value="true">true</option>
                  <option value="false">false</option>
                </select>
              </Field>
              <Field label="词条 ID">
                <input className={`${controlClassName} opacity-70`} value={editID || '创建后生成'} disabled />
              </Field>
            </div>
            <Field label="备注">
              <textarea
                className={`${fieldClassName} min-h-32 resize-y px-4 py-3`}
                value={form.notes}
                onChange={(event) => setForm((prev) => ({ ...prev, notes: event.target.value }))}
              />
            </Field>
            <button className={primaryButtonClassName} type="submit">
              <CheckCircle2 className="size-4" />
              保存词条
            </button>
          </form>
        )}
        <p className="mt-4 text-sm text-slate-400">{status}</p>
      </section>
    </AdminScaffold>
  );
}

function KeywordImportPage() {
  useDocumentTitle('givezj8.cn · Excel 导入');
  const [file, setFile] = useState<File | null>(null);
  const [status, setStatus] = useState('等待上传');
  const [result, setResult] = useState<KeywordSeedImportResponse | null>(null);

  async function handleUpload() {
    if (!file) {
      setStatus('请选择 .xlsx 文件');
      return;
    }
    setStatus('上传中');
    try {
      const payload = await importKeywordSeeds(file);
      setResult(payload);
      setStatus(`导入完成：成功 ${payload.imported_count}，失败 ${payload.failed_count}`);
    } catch (err) {
      setStatus(err instanceof APIError ? err.message : '导入失败');
      setResult(null);
    }
  }

  return (
    <AdminScaffold
      title="Excel 导入"
      description="支持下载模板并上传工作簿，后台仍使用原有 FormData 导入接口。"
      toolbar={
        <>
          <NavLink href="/admin">后台首页</NavLink>
          <NavLink href="/admin/keywords">词库列表</NavLink>
        </>
      }
    >
      <div className="grid gap-5 lg:grid-cols-[1.15fr_0.85fr]">
        <section className="rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-6 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)]">
          <h3 className="text-lg font-semibold text-slate-100">上传工作簿</h3>
          <p className="mt-3 text-sm leading-6 text-slate-400">
            支持下载模板并上传 `.xlsx` 文件。后端会直接解析第一张表并按共享文档字段校验。
          </p>
          <label className="mt-6 flex min-h-44 cursor-pointer flex-col items-center justify-center rounded-[24px] border border-dashed border-cyan-300/30 bg-cyan-400/5 p-6 text-center">
            <Upload className="size-8 text-cyan-300" />
            <span className="mt-3 text-sm text-slate-200">{file ? file.name : '点击选择 Excel 文件'}</span>
            <span className="mt-1 text-xs text-slate-500">仅支持 .xlsx</span>
            <input className="hidden" type="file" accept=".xlsx" onChange={(event) => setFile(event.target.files?.[0] ?? null)} />
          </label>
          <div className="mt-6 flex flex-wrap gap-3">
            <NavLink href="/admin/api/v1/keyword-seeds/template">
              <Download className="size-4" />
              下载模板
            </NavLink>
            <button className={primaryButtonClassName} type="button" onClick={() => void handleUpload()}>
              <Upload className="size-4" />
              开始导入
            </button>
          </div>
          <p className="mt-4 text-sm text-slate-400">{status}</p>
        </section>

        <aside className="rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-6 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)]">
          <h3 className="text-lg font-semibold text-slate-100">导入结果</h3>
          <pre className="mt-4 overflow-x-auto whitespace-pre-wrap break-words rounded-2xl border border-cyan-300/12 bg-slate-900/74 p-4 text-xs leading-6 text-slate-400">
            {result ? JSON.stringify(result, null, 2) : '等待上传'}
          </pre>
        </aside>
      </div>
    </AdminScaffold>
  );
}

function JDSchedulePage() {
  useDocumentTitle('givezj8.cn · JD 定时采集');
  const [form, setForm] = useState<CollectorScheduleUpsertRequest>({
    enabled: false,
    schedule_time: '03:00',
    request_interval_seconds: 3,
    query_limit: 5,
  });
  const [configured, setConfigured] = useState(false);
  const [status, setStatus] = useState('读取中');
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    getJDScheduleConfig()
      .then((payload) => {
        if (cancelled) {
          return;
        }
        setConfigured(payload.configured);
        if (payload.config) {
          setForm({
            enabled: payload.config.enabled,
            schedule_time: payload.config.schedule_time,
            request_interval_seconds: payload.config.request_interval_seconds,
            query_limit: payload.config.query_limit,
          });
          setStatus(payload.config.enabled ? '当前已启用定时采集' : '当前已配置但未启用');
        } else {
          setStatus('当前未配置，不会自动启动定时采集');
        }
      })
      .catch((err) => {
        if (err instanceof APIError && err.status === 401) {
          window.location.href = '/admin/login';
          return;
        }
        setStatus(err instanceof APIError ? err.message : '读取配置失败');
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    try {
      const payload = await updateJDScheduleConfig({
        enabled: form.enabled,
        schedule_time: form.schedule_time,
        request_interval_seconds: Number(form.request_interval_seconds),
        query_limit: Number(form.query_limit),
      });
      setConfigured(payload.configured);
      if (payload.config) {
        setForm({
          enabled: payload.config.enabled,
          schedule_time: payload.config.schedule_time,
          request_interval_seconds: payload.config.request_interval_seconds,
          query_limit: payload.config.query_limit,
        });
      }
      setStatus(payload.config?.enabled ? '定时采集已启用并保存' : '定时采集配置已保存，但当前关闭');
    } catch (err) {
      setStatus(err instanceof APIError ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  }

  return (
    <AdminScaffold
      title="JD 定时采集"
      description="由后台统一管理每日采集时间。没有配置或配置关闭时，jd-collector 不会启动定时任务。"
      toolbar={
        <>
          <NavLink href="/admin">后台首页</NavLink>
          <NavLink href="/admin/keywords">词库列表</NavLink>
        </>
      }
    >
      <div className="grid gap-5 lg:grid-cols-[1.1fr_0.9fr]">
        <section className="rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-6 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)]">
          {loading ? (
            <p className="text-sm text-slate-400">读取配置中...</p>
          ) : (
            <form className="space-y-5" onSubmit={handleSubmit}>
              <div className="flex items-center justify-between gap-4 rounded-2xl border border-cyan-300/18 bg-slate-900/72 px-4 py-4">
                <div>
                  <p className="text-sm font-semibold text-slate-100">开启定时任务</p>
                  <p className="mt-1 text-sm leading-6 text-slate-400">关闭后，服务不会按天自动采集；仍可保留配置。</p>
                </div>
                <button
                  type="button"
                  className={`inline-flex h-10 min-w-24 items-center justify-center rounded-full border px-4 text-sm font-medium transition ${
                    form.enabled
                      ? 'border-cyan-200/60 bg-cyan-300/18 text-cyan-100'
                      : 'border-cyan-300/18 bg-slate-950/70 text-slate-300'
                  }`}
                  onClick={() => setForm((prev) => ({ ...prev, enabled: !prev.enabled }))}
                >
                  {form.enabled ? '已开启' : '已关闭'}
                </button>
              </div>

              <div className="grid gap-5 md:grid-cols-2">
                <Field label="每日执行时间">
                  <input
                    className={controlClassName}
                    type="time"
                    value={form.schedule_time}
                    onChange={(event) => setForm((prev) => ({ ...prev, schedule_time: event.target.value }))}
                    required
                  />
                </Field>
                <Field label="每次查询条数">
                  <input
                    className={controlClassName}
                    type="number"
                    min="1"
                    step="1"
                    value={form.query_limit}
                    onChange={(event) => setForm((prev) => ({ ...prev, query_limit: Number(event.target.value) }))}
                    required
                  />
                </Field>
              </div>

              <Field label="每次接口请求间隔（秒）">
                <input
                  className={controlClassName}
                  type="number"
                  min="0"
                  step="1"
                  value={form.request_interval_seconds}
                  onChange={(event) =>
                    setForm((prev) => ({ ...prev, request_interval_seconds: Number(event.target.value) }))
                  }
                  required
                />
              </Field>

              <button className={primaryButtonClassName} type="submit" disabled={saving}>
                <CheckCircle2 className="size-4" />
                {saving ? '保存中...' : '保存调度配置'}
              </button>
            </form>
          )}
          <p className="mt-4 text-sm text-slate-400">{status}</p>
        </section>

        <aside className="rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-6 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)]">
          <h3 className="text-lg font-semibold text-slate-100">当前规则</h3>
          <div className="mt-4 space-y-3 text-sm leading-6 text-slate-400">
            <p>1. 没有配置时，jd-collector 不会自动启动定时采集。</p>
            <p>2. 配置已保存但关闭时，服务继续运行，但不会进入每日采集。</p>
            <p>3. 开启后按每日时间执行，并按设置的接口间隔串行请求。</p>
            <p>4. 采集完成后会写入原始商品、价格快照和型号级每日价格快照。</p>
          </div>
          <div className="mt-5 rounded-2xl border border-cyan-300/18 bg-slate-900/72 p-4 text-sm text-slate-300">
            当前状态：{configured ? '已存在配置' : '未配置'}
          </div>
        </aside>
      </div>
    </AdminScaffold>
  );
}

function ResultPanel({
  result,
  summaryText,
  selectedItems,
  error,
}: {
  result: CatalogRecommendationResponse | null;
  summaryText: string;
  selectedItems: CatalogRecommendationResponse['selection']['selected_items'];
  error: string;
}) {
  if (!result) {
    return (
      <div className="app-result-panel rounded-[24px] border border-cyan-300/24 bg-slate-950/82 p-6 shadow-[0_16px_48px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)]">
        <div className="flex min-h-[500px] flex-col items-center justify-center px-6 py-10 text-center">
          <div className="mx-auto mb-5 flex h-16 w-16 items-center justify-center rounded-2xl border border-cyan-300/24 bg-cyan-400/8 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.08)] sm:mb-6 sm:h-20 sm:w-20">
            <Sparkles className="size-8 text-cyan-300 sm:size-10" />
          </div>
          <h3 className="mb-2 text-lg font-semibold text-slate-50 sm:text-xl">先生成一套靠谱方案</h3>
          <p className="mx-auto max-w-sm text-sm leading-6 text-slate-400 sm:text-base">
            左侧填完后，这里会展示总价、推荐摘要、配件清单和每个配件的价格依据。
          </p>
          {error ? <p className="mt-4 text-sm text-rose-300">{error}</p> : null}
        </div>
      </div>
    );
  }

  const advice = result.advice;
  return (
    <div className="app-result-panel animate-in fade-in slide-in-from-bottom-4 rounded-[24px] border border-cyan-300/24 bg-slate-950/82 p-6 duration-500 shadow-[0_16px_48px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)]">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300/72">推荐结果</p>
          <h3 className="mt-2 text-2xl font-semibold text-slate-50">这是一套可以直接看的装机方案</h3>
          <p className="mt-2 text-sm text-slate-400">先看总价和摘要，再决定要不要细看每个配件。</p>
        </div>
        <StatusPill>{result.request_status.cache_hit ? '已返回最近结果' : '已生成新结果'}</StatusPill>
      </div>

      <section className="mt-5 grid gap-4 sm:grid-cols-3">
        <HighlightCard
          label="估算总价"
          value={`¥${result.selection.estimated_total.toLocaleString()}`}
          hint="当前这套配置按选中配件价格估算"
        />
        <HighlightCard
          label="已选配件"
          value={`${selectedItems.length} 项`}
          hint="包含 CPU、GPU、主板、内存等核心部件"
        />
        <HighlightCard
          label="价格样本"
          value={`${result.catalog_item_count} 条`}
          hint="当前推荐建立在已整理的价格目录之上"
        />
      </section>

      <section className="app-soft-panel mt-5 rounded-xl border border-cyan-300/24 bg-cyan-400/4 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-cyan-300/72">推荐摘要</p>
        <p className="mt-2 text-sm leading-7 text-slate-200">{summaryText}</p>
      </section>

      {result.selection.warnings?.length ? (
        <section className="mt-4 rounded-xl border border-amber-300/24 bg-amber-400/8 p-4 text-sm text-amber-100">
          <p className="font-semibold">你需要先注意</p>
          <div className="mt-2 space-y-1">
            {result.selection.warnings.map((item) => (
              <p key={item}>• {item}</p>
            ))}
          </div>
        </section>
      ) : null}

      <section className="mt-6">
        <div className="flex items-center justify-between">
          <div>
            <h4 className="text-sm font-semibold text-slate-100">配件清单</h4>
            <p className="mt-1 text-xs text-slate-500">每项都附带参考价和选择依据。</p>
          </div>
          <span className="text-xs text-slate-500">共 {selectedItems.length} 项</span>
        </div>
        <div className="mt-4 space-y-3">
          {selectedItems.map((item) => (
            <article
              key={`${item.category}-${item.normalized_key}`}
              className="app-soft-panel rounded-xl border border-cyan-300/20 bg-slate-900/72 px-4 py-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]"
            >
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="size-4 shrink-0 text-cyan-300" />
                    <p className="text-sm font-medium text-slate-100">{item.category}</p>
                  </div>
                  <p className="mt-1 break-words text-sm text-slate-300">{item.display_name}</p>
                  <div className="mt-2 flex flex-wrap gap-2 text-xs text-slate-500">
                    <span className="rounded-full border border-cyan-300/14 bg-cyan-400/6 px-2 py-1">
                      参考价 ¥{item.selected_price.toLocaleString()}
                    </span>
                    <span className="rounded-full border border-cyan-300/14 bg-white/5 px-2 py-1">
                      中位价 ¥{item.median_price.toLocaleString()}
                    </span>
                    <span className="rounded-full border border-cyan-300/14 bg-white/5 px-2 py-1">
                      样本 {item.sample_count}
                    </span>
                  </div>
                  {item.reasons?.length ? (
                    <div className="mt-2 space-y-1">
                      {item.reasons.map((reason) => (
                        <p className="text-xs leading-5 text-slate-400" key={reason}>
                          • {reason}
                        </p>
                      ))}
                    </div>
                  ) : null}
                </div>
                <div className="shrink-0 text-left sm:text-right">
                  <p className="text-xs uppercase tracking-[0.18em] text-slate-500">当前参考价</p>
                  <p className="mt-1 text-xl font-semibold text-cyan-200">¥{item.selected_price.toLocaleString()}</p>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>

      {advice ? (
        <section className="mt-6 grid gap-3 md:grid-cols-2">
          <AdviceCard title="为什么这样配" items={advice.reasons} />
          <AdviceCard title="适合什么人" items={advice.fit_for} />
          <AdviceCard title="你需要注意" items={advice.risks} />
          <AdviceCard title="后续还能怎么升" items={advice.upgrade_advice} />
        </section>
      ) : null}

      {advice?.alternative_note ? (
        <section className="app-soft-panel mt-4 rounded-xl border border-cyan-300/12 bg-slate-900/72 p-4 text-sm leading-6 text-slate-400">
          {advice.alternative_note}
        </section>
      ) : null}
    </div>
  );
}

function AdviceCard({ title, items }: { title: string; items?: string[] }) {
  return (
    <div className="app-soft-panel rounded-xl border border-cyan-300/20 bg-slate-900/72 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]">
      <p className="text-sm font-semibold text-slate-100">{title}</p>
      <div className="mt-3 space-y-2 text-sm leading-6 text-slate-400">
        {items?.length ? items.map((item) => <p key={item}>• {item}</p>) : <p>暂无内容</p>}
      </div>
    </div>
  );
}

function AdminScaffold({
  title,
  description,
  toolbar,
  children,
}: {
  title: string;
  description: string;
  toolbar?: ReactNode;
  children: ReactNode;
}) {
  const { isLight, toggleTheme } = useThemeMode();
  return (
    <div className="app-root app-admin-shell relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_rgba(55,197,255,0.12),_transparent_28%),linear-gradient(180deg,_rgba(5,12,22,0.92),_rgba(5,12,22,1))]">
      <div className="pointer-events-none absolute inset-0 opacity-55 [background-image:linear-gradient(rgba(109,184,255,0.12)_1px,transparent_1px),linear-gradient(90deg,rgba(109,184,255,0.12)_1px,transparent_1px)] [background-size:36px_36px]" />
      <div className="container relative z-10 mx-auto max-w-7xl px-4 py-8 sm:py-10">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300/72">givezj8.cn / admin</p>
            <h1 className="mt-2 text-4xl leading-none text-slate-50 sm:text-5xl">{title}</h1>
            <p className="mt-4 max-w-3xl text-sm leading-7 text-slate-400">{description}</p>
          </div>
          <div className="flex flex-wrap gap-2">
            <ThemeToggleButton isLight={isLight} onToggle={toggleTheme} />
            {toolbar}
          </div>
        </div>
        <div className="mt-8">{children}</div>
      </div>
    </div>
  );
}

function AdminFeatureCard({
  href,
  icon,
  title,
  text,
}: {
  href: string;
  icon: ReactNode;
  title: string;
  text: string;
}) {
  return (
    <a
      className="app-soft-panel group rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-5 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)] transition hover:-translate-y-0.5 hover:border-cyan-200/40"
      href={href}
    >
      <div className="flex items-center gap-3">
        <div className="rounded-xl border border-cyan-300/18 bg-cyan-400/10 p-3">{icon}</div>
        <ArrowRight className="size-4 text-slate-500 transition group-hover:text-cyan-200" />
      </div>
      <p className="mt-5 text-lg font-semibold text-slate-100">{title}</p>
      <p className="mt-2 text-sm leading-6 text-slate-400">{text}</p>
    </a>
  );
}

function HighlightCard({ label, value, hint }: { label: string; value: string; hint: string }) {
  return (
    <div className="app-soft-panel rounded-2xl border border-cyan-300/20 bg-slate-900/74 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]">
      <p className="text-xs font-semibold uppercase tracking-[0.18em] text-cyan-300/72">{label}</p>
      <p className="mt-3 text-2xl font-semibold text-slate-50">{value}</p>
      <p className="mt-2 text-sm leading-6 text-slate-400">{hint}</p>
    </div>
  );
}

function StatusPill({ children }: { children: ReactNode }) {
  return (
    <span className="app-status-pill inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-3 py-2 text-xs text-slate-300">
      {children}
    </span>
  );
}

function NavLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      className="app-nav-link inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-4 py-2 text-sm text-slate-300 transition hover:border-cyan-200/40 hover:text-slate-50"
      href={href}
    >
      {children}
    </a>
  );
}

function ThemeToggleButton({ isLight, onToggle }: { isLight: boolean; onToggle: () => void }) {
  return (
    <button
      type="button"
      className="app-theme-toggle inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-4 py-2 text-sm text-slate-300 transition hover:border-cyan-200/40 hover:text-slate-50"
      onClick={onToggle}
    >
      {isLight ? <Moon className="size-4" /> : <Sun className="size-4" />}
      <span>{isLight ? '夜间' : '日间'}</span>
    </button>
  );
}

function Field({ label, icon, children }: { label: string; icon?: ReactNode; children: ReactNode }) {
  return (
    <label className="block space-y-2">
      <span className="flex items-center gap-2 text-sm text-slate-200">
        {icon}
        {label}
      </span>
      {children}
    </label>
  );
}

const fieldClassName =
  'app-control w-full rounded-xl border border-cyan-300/18 bg-slate-900/72 text-sm text-slate-100 outline-none transition placeholder:text-slate-500 focus:border-cyan-200/50';

const controlClassName = `${fieldClassName} h-12 px-4`;
const scrollPanelClassName =
  'app-scroll-panel space-y-4 lg:h-full lg:overflow-auto lg:pr-2 [scrollbar-width:thin] [scrollbar-color:rgba(103,232,249,0.28)_transparent] [&::-webkit-scrollbar]:w-2 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-cyan-300/25';

const secondaryButtonClassName =
  'inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-4 py-2 text-sm text-slate-300 transition hover:border-cyan-200/40 hover:text-slate-50';

const primaryButtonClassName =
  'inline-flex items-center gap-2 rounded-xl bg-cyan-300 px-4 py-3 text-sm font-semibold text-slate-950 transition hover:bg-cyan-200';

function useThemeMode() {
  const [themeMode, setThemeMode] = useState<'dark' | 'light'>(() => {
    if (typeof window === 'undefined') {
      return 'dark';
    }
    return window.localStorage.getItem(themeStorageKey) === 'light' ? 'light' : 'dark';
  });

  useEffect(() => {
    const root = document.documentElement;
    root.classList.toggle('light-theme', themeMode === 'light');
    window.localStorage.setItem(themeStorageKey, themeMode);
  }, [themeMode]);

  return {
    isLight: themeMode === 'light',
    toggleTheme: () => setThemeMode((prev) => (prev === 'light' ? 'dark' : 'light')),
  };
}

function useDocumentTitle(title: string) {
  useEffect(() => {
    document.title = title;
  }, [title]);
}
