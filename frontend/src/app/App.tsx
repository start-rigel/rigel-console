import { FormEvent, ReactNode, useEffect, useMemo, useState } from 'react';
import {
  ArrowRight,
  CheckCircle2,
  Database,
  Download,
  FileSpreadsheet,
  Gauge,
  Info,
  LogOut,
  MonitorSmartphone,
  Search,
  Settings2,
  ShieldCheck,
  Sparkles,
  Upload,
  Wallet,
} from 'lucide-react';
import {
  APIError,
  createKeywordSeed,
  generateRecommendation,
  getAnonymousSession,
  getKeywordSeed,
  importKeywordSeeds,
  listKeywordSeeds,
  loginAdmin,
  logoutAdmin,
  setKeywordSeedEnabled,
  updateKeywordSeed,
} from './lib/api';
import type {
  CatalogRecommendationResponse,
  GenerateBuildRequest,
  KeywordSeed,
  KeywordSeedImportResponse,
  KeywordSeedUpsertRequest,
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
  if (editMatch) {
    return <KeywordFormPage editID={decodeURIComponent(editMatch[1])} />;
  }
  return <RecommendationPage />;
}

function RecommendationPage() {
  useDocumentTitle('givezj8.cn · 匿名装机推荐');
  const [form, setForm] = useState<PublicFormState>(defaultPublicForm);
  const [anonymousID, setAnonymousID] = useState('');
  const [remaining, setRemaining] = useState('-');
  const [cooldown, setCooldown] = useState('0 秒');
  const [sessionStatus, setSessionStatus] = useState('匿名会话初始化中');
  const [requestStatus, setRequestStatus] = useState('等待生成');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  const [result, setResult] = useState<CatalogRecommendationResponse | null>(null);

  useEffect(() => {
    let cancelled = false;
    getAnonymousSession()
      .then((session) => {
        if (cancelled) {
          return;
        }
        setAnonymousID(session.anonymous_id);
        setRemaining(String(session.remaining_ai_requests ?? '-'));
        setCooldown(`${session.cooldown_seconds ?? 0} 秒`);
        setSessionStatus(session.anonymous_id ? `匿名会话 ${session.anonymous_id}` : '匿名会话获取失败');
      })
      .catch(() => {
        if (!cancelled) {
          setSessionStatus('匿名会话初始化失败');
        }
      });
    return () => {
      cancelled = true;
    };
  }, []);

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
      setCooldown(`${data.request_status.cooldown_seconds ?? 0} 秒`);
      setRequestStatus(data.request_status.cache_hit ? '已命中缓存' : '已生成新结果');
    } catch (err) {
      const message = err instanceof APIError ? err.message : '生成推荐失败';
      const cooldownSeconds = err instanceof APIError ? err.cooldownSeconds : 0;
      setError(message);
      setCooldown(`${cooldownSeconds} 秒`);
      setRequestStatus(message);
    } finally {
      setIsLoading(false);
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
    return '价格目录返回后，这里会展示总价、推荐摘要和每个配件的选择依据。';
  }, [advice?.summary, result?.catalog_warnings]);

  return (
    <div className="relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_rgba(55,197,255,0.12),_transparent_28%),linear-gradient(180deg,_rgba(5,12,22,0.92),_rgba(5,12,22,1))]">
      <div className="pointer-events-none absolute inset-0 opacity-55 [background-image:linear-gradient(rgba(109,184,255,0.12)_1px,transparent_1px),linear-gradient(90deg,rgba(109,184,255,0.12)_1px,transparent_1px)] [background-size:36px_36px]" />
      <div className="pointer-events-none absolute inset-x-0 top-0 h-80 bg-[radial-gradient(circle_at_top,rgba(95,224,255,0.22),transparent_55%)]" />

      <header className="sticky top-0 z-20 border-b border-cyan-400/10 bg-slate-950/70 backdrop-blur-md">
        <div className="container mx-auto flex items-center justify-between gap-4 px-4 py-4">
          <div className="flex min-w-0 items-center gap-3">
            <div className="rounded-lg border border-cyan-400/20 bg-cyan-400/10 p-2 shadow-[0_0_24px_rgba(95,224,255,0.12)]">
              <Sparkles className="size-5 text-white sm:size-6" />
            </div>
            <div className="min-w-0">
              <p className="text-[10px] uppercase tracking-[0.34em] text-cyan-300/70 sm:text-xs">PC Configuration Console</p>
              <h1 className="text-xl font-bold text-slate-50 sm:text-2xl">给我装机吧</h1>
              <p className="text-xs text-slate-400 sm:text-sm">预算分配 · 性能取舍 · 配置解释</p>
            </div>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <StatusPill>{sessionStatus}</StatusPill>
            <NavLink href="/admin/login">后台管理</NavLink>
          </div>
        </div>
      </header>

      <main className="container mx-auto px-4 py-5 sm:py-8">
        <section className="mx-auto mb-8 max-w-7xl rounded-[28px] border border-cyan-300/28 bg-slate-950/78 p-5 shadow-[0_24px_80px_rgba(0,0,0,0.28),inset_0_0_0_1px_rgba(130,220,255,0.08)] backdrop-blur sm:mb-10 sm:p-7">
          <div className="grid gap-6 lg:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)] lg:items-start">
            <div className="space-y-6">
              <section>
                <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300/72">匿名直接使用</p>
                <h2 className="mt-3 max-w-3xl text-4xl leading-[0.92] text-slate-50 sm:text-5xl lg:text-6xl">
                  用新界面承接真实价格目录和 AI 推荐结果。
                </h2>
                <p className="mt-4 max-w-2xl text-sm leading-7 text-slate-400 sm:text-base">
                  界面采用新的 React 视觉体系，但推荐请求、匿名配额、缓存命中和后台词库功能仍然接在项目原有接口上。
                </p>
              </section>

              <div className="grid gap-4 sm:grid-cols-3">
                <MetricCard icon={<Wallet className="size-4 text-cyan-300" />} label="剩余匿名 AI 次数" value={remaining} />
                <MetricCard icon={<Gauge className="size-4 text-cyan-300" />} label="冷却时间" value={cooldown} />
                <MetricCard
                  icon={<ShieldCheck className="size-4 text-cyan-300" />}
                  label="请求状态"
                  value={requestStatus}
                />
              </div>

              <section className="rounded-[24px] border border-cyan-300/24 bg-slate-950/82 p-5 shadow-[0_16px_48px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)]">
                <div className="mb-6">
                  <h3 className="text-lg font-semibold text-slate-50 sm:text-xl">告诉我你的预算和用途</h3>
                  <p className="mt-2 text-sm leading-6 text-slate-400">
                    主要交互沿用目标页面布局，同时补上品牌偏好、特殊要求和补充说明，让原有推荐能力继续可用。
                  </p>
                </div>

                <form onSubmit={handleSubmit} className="space-y-5 sm:space-y-6">
                  <Field label="预算 (¥)" icon={<Wallet className="size-4" />}>
                    <input
                      className={controlClassName}
                      type="number"
                      min="1000"
                      step="100"
                      value={form.budget}
                      onChange={(event) => setForm((prev) => ({ ...prev, budget: event.target.value }))}
                      required
                    />
                  </Field>

                  <div className="grid gap-5 sm:grid-cols-2">
                    <Field label="用途" icon={<MonitorSmartphone className="size-4" />}>
                      <select
                        className={controlClassName}
                        value={form.useCase}
                        onChange={(event) => setForm((prev) => ({ ...prev, useCase: event.target.value }))}
                      >
                        <option value="gaming">游戏</option>
                        <option value="office">办公</option>
                        <option value="design">设计 / 剪辑</option>
                      </select>
                    </Field>
                    <Field label="装机模式" icon={<Settings2 className="size-4" />}>
                      <select
                        className={controlClassName}
                        value={form.buildMode}
                        onChange={(event) => setForm((prev) => ({ ...prev, buildMode: event.target.value }))}
                      >
                        <option value="mixed">混合来源</option>
                        <option value="new_only">只看全新</option>
                        <option value="used_only">只看二手</option>
                      </select>
                    </Field>
                  </div>

                  <div className="rounded-xl border border-cyan-300/24 bg-cyan-400/5 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
                    <p className="text-sm font-medium text-slate-100">高级条件</p>
                    <p className="mt-1 text-sm leading-6 text-slate-400">这里补齐原项目已有的品牌偏好、特殊要求和补充说明。</p>
                    <div className="mt-4 grid gap-4 sm:grid-cols-2">
                      <Field label="CPU 品牌偏好">
                        <input
                          className={controlClassName}
                          placeholder="AMD / Intel"
                          value={form.cpuBrand}
                          onChange={(event) => setForm((prev) => ({ ...prev, cpuBrand: event.target.value }))}
                        />
                      </Field>
                      <Field label="GPU 品牌偏好">
                        <input
                          className={controlClassName}
                          placeholder="NVIDIA / AMD"
                          value={form.gpuBrand}
                          onChange={(event) => setForm((prev) => ({ ...prev, gpuBrand: event.target.value }))}
                        />
                      </Field>
                    </div>
                    <div className="mt-4 space-y-4">
                      <Field label="特殊要求">
                        <input
                          className={controlClassName}
                          placeholder="wifi_motherboard, low_noise"
                          value={form.specialRequirements}
                          onChange={(event) => setForm((prev) => ({ ...prev, specialRequirements: event.target.value }))}
                        />
                      </Field>
                      <Field label="补充说明">
                        <textarea
                          className={`${fieldClassName} min-h-28 resize-y px-4 py-3`}
                          placeholder="例如：1080p 游戏为主，希望价格稳一点。"
                          value={form.notes}
                          onChange={(event) => setForm((prev) => ({ ...prev, notes: event.target.value }))}
                        />
                      </Field>
                    </div>
                  </div>

                  <button
                    type="submit"
                    className="flex h-12 w-full items-center justify-center rounded-xl bg-cyan-300 text-sm font-semibold text-slate-950 transition hover:bg-cyan-200 disabled:cursor-not-allowed disabled:bg-cyan-100"
                    disabled={isLoading}
                  >
                    {isLoading ? '生成中...' : '生成配置方案'}
                  </button>
                </form>
              </section>

              <section className="rounded-2xl border border-cyan-300/24 bg-slate-950/82 p-5 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.06)]">
                <div className="flex items-center gap-2 text-slate-100">
                  <Info className="size-4 text-cyan-300" />
                  <p className="font-semibold">部署原则</p>
                </div>
                <div className="mt-3 space-y-2 text-sm text-slate-400">
                  <p>• 价格目录来自现有后端链路，不再使用本地模拟数据。</p>
                  <p>• 相同请求优先命中缓存，不重复消耗匿名 AI 次数。</p>
                  <p>• 后台词库页已统一到这套 React 视觉体系，但仍走原有管理接口。</p>
                </div>
              </section>
            </div>

            <section>
              <ResultPanel result={result} summaryText={summaryText} selectedItems={selectedItems} error={error} />
            </section>
          </div>
        </section>

        <section className="mx-auto mt-14 max-w-7xl rounded-[24px] border border-cyan-300/24 bg-slate-950/78 p-5 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)] sm:p-6">
          <div className="grid gap-4 md:grid-cols-3">
            <InfoBlock title="你会拿到什么" text="一套真实的型号级价格清单结果，以及 AI 给出的结构化建议、风险和升级方向。" />
            <InfoBlock title="建议怎么读" text="先看估算总价和警告，再看建议摘要，最后逐项确认 CPU、GPU、主板与存储的取舍。" />
            <InfoBlock title="后台还能做什么" text="管理员仍然可以登录后台，维护词库、导入 Excel、启停词条和导出当前配置词库。" />
          </div>
        </section>
      </main>

      <footer className="mt-16 border-t border-cyan-400/10 bg-slate-950/70 backdrop-blur-sm">
        <div className="container mx-auto px-4 py-6 text-center text-sm text-slate-500">
          <p>© 2026 givezj8.cn · 新前端接入真实推荐链路</p>
        </div>
      </footer>
    </div>
  );
}

function AdminLoginPage() {
  useDocumentTitle('givezj8.cn · 后台登录');
  const [username, setUsername] = useState('admin');
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
      <div className="rounded-[24px] border border-cyan-300/24 bg-slate-950/82 p-6 shadow-[0_16px_48px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)]">
        <div className="flex min-h-[500px] flex-col items-center justify-center px-6 py-10 text-center">
          <div className="mx-auto mb-5 flex h-16 w-16 items-center justify-center rounded-2xl border border-cyan-300/24 bg-cyan-400/8 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.08)] sm:mb-6 sm:h-20 sm:w-20">
            <Sparkles className="size-8 text-cyan-300 sm:size-10" />
          </div>
          <h3 className="mb-2 text-lg font-semibold text-slate-50 sm:text-xl">先生成一套靠谱方案</h3>
          <p className="mx-auto max-w-sm text-sm leading-6 text-slate-400 sm:text-base">
            左侧填完后，这里会展示目录条目数、估算总价、顾问摘要和具体选中的配件列表。
          </p>
          {error ? <p className="mt-4 text-sm text-rose-300">{error}</p> : null}
        </div>
      </div>
    );
  }

  const advice = result.advice;
  return (
    <div className="animate-in fade-in slide-in-from-bottom-4 rounded-[24px] border border-cyan-300/24 bg-slate-950/82 p-6 duration-500 shadow-[0_16px_48px_rgba(0,0,0,0.26),inset_0_0_0_1px_rgba(130,220,255,0.06)]">
      <div className="flex flex-wrap items-start justify-between gap-4">
        <div>
          <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300/72">推荐结果</p>
          <h3 className="mt-2 text-2xl font-semibold text-slate-50">目录条目 {result.catalog_item_count}</h3>
          <p className="mt-2 text-sm text-slate-400">估算总价 ¥{result.selection.estimated_total.toLocaleString()}</p>
        </div>
        <StatusPill>{result.request_status.cache_hit ? '已返回最近结果' : '已生成新结果'}</StatusPill>
      </div>

      <section className="mt-5 rounded-xl border border-cyan-300/24 bg-cyan-400/4 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-cyan-300/72">顾问摘要</p>
        <p className="mt-2 text-sm leading-7 text-slate-200">{summaryText}</p>
      </section>

      {result.selection.warnings?.length ? (
        <section className="mt-4 rounded-xl border border-amber-300/24 bg-amber-400/8 p-4 text-sm text-amber-100">
          <p className="font-semibold">当前警告</p>
          <div className="mt-2 space-y-1">
            {result.selection.warnings.map((item) => (
              <p key={item}>• {item}</p>
            ))}
          </div>
        </section>
      ) : null}

      <section className="mt-6">
        <div className="flex items-center justify-between">
          <h4 className="text-sm font-semibold text-slate-100">主机配置</h4>
          <span className="text-xs text-slate-500">共 {selectedItems.length} 项</span>
        </div>
        <div className="mt-4 space-y-3">
          {selectedItems.map((item) => (
            <article
              key={`${item.category}-${item.normalized_key}`}
              className="rounded-xl border border-cyan-300/20 bg-slate-900/72 px-4 py-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]"
            >
              <div className="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
                <div className="min-w-0 flex-1">
                  <div className="flex items-center gap-2">
                    <CheckCircle2 className="size-4 shrink-0 text-cyan-300" />
                    <p className="text-sm font-medium text-slate-100">{item.category}</p>
                  </div>
                  <p className="mt-1 break-words text-sm text-slate-300">{item.display_name}</p>
                  <p className="mt-1 text-xs text-slate-500">
                    样本 {item.sample_count} · 中位价 ¥{item.median_price.toLocaleString()}
                  </p>
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
                  <p className="font-semibold text-cyan-200">¥{item.selected_price.toLocaleString()}</p>
                </div>
              </div>
            </article>
          ))}
        </div>
      </section>

      {advice ? (
        <section className="mt-6 grid gap-3 md:grid-cols-2">
          <AdviceCard title="推荐理由" items={advice.reasons} />
          <AdviceCard title="适用场景" items={advice.fit_for} />
          <AdviceCard title="风险提醒" items={advice.risks} />
          <AdviceCard title="升级建议" items={advice.upgrade_advice} />
        </section>
      ) : null}

      {advice?.alternative_note ? (
        <section className="mt-4 rounded-xl border border-cyan-300/12 bg-slate-900/72 p-4 text-sm leading-6 text-slate-400">
          {advice.alternative_note}
        </section>
      ) : null}
    </div>
  );
}

function AdviceCard({ title, items }: { title: string; items?: string[] }) {
  return (
    <div className="rounded-xl border border-cyan-300/20 bg-slate-900/72 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]">
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
  return (
    <div className="relative min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top,_rgba(55,197,255,0.12),_transparent_28%),linear-gradient(180deg,_rgba(5,12,22,0.92),_rgba(5,12,22,1))]">
      <div className="pointer-events-none absolute inset-0 opacity-55 [background-image:linear-gradient(rgba(109,184,255,0.12)_1px,transparent_1px),linear-gradient(90deg,rgba(109,184,255,0.12)_1px,transparent_1px)] [background-size:36px_36px]" />
      <div className="container relative z-10 mx-auto max-w-7xl px-4 py-8 sm:py-10">
        <div className="flex flex-wrap items-start justify-between gap-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-[0.24em] text-cyan-300/72">givezj8.cn / admin</p>
            <h1 className="mt-2 text-4xl leading-none text-slate-50 sm:text-5xl">{title}</h1>
            <p className="mt-4 max-w-3xl text-sm leading-7 text-slate-400">{description}</p>
          </div>
          {toolbar ? <div className="flex flex-wrap gap-2">{toolbar}</div> : null}
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
      className="group rounded-[24px] border border-cyan-300/24 bg-slate-950/84 p-5 shadow-[0_18px_56px_rgba(0,0,0,0.22),inset_0_0_0_1px_rgba(130,220,255,0.05)] transition hover:-translate-y-0.5 hover:border-cyan-200/40"
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

function MetricCard({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <div className="rounded-2xl border border-cyan-300/20 bg-slate-950/82 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
      <div className="flex items-center gap-2 text-sm text-slate-400">
        {icon}
        <span>{label}</span>
      </div>
      <p className="mt-4 text-2xl font-semibold text-slate-50">{value}</p>
    </div>
  );
}

function InfoBlock({ title, text }: { title: string; text: string }) {
  return (
    <div>
      <p className="text-sm font-semibold text-slate-100">{title}</p>
      <p className="mt-2 text-sm leading-6 text-slate-400">{text}</p>
    </div>
  );
}

function StatusPill({ children }: { children: ReactNode }) {
  return (
    <span className="inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-3 py-2 text-xs text-slate-300">
      {children}
    </span>
  );
}

function NavLink({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      className="inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-4 py-2 text-sm text-slate-300 transition hover:border-cyan-200/40 hover:text-slate-50"
      href={href}
    >
      {children}
    </a>
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
  'w-full rounded-xl border border-cyan-300/18 bg-slate-900/72 text-sm text-slate-100 outline-none transition placeholder:text-slate-500 focus:border-cyan-200/50';

const controlClassName = `${fieldClassName} h-12 px-4`;

const secondaryButtonClassName =
  'inline-flex items-center gap-2 rounded-full border border-cyan-300/18 bg-white/5 px-4 py-2 text-sm text-slate-300 transition hover:border-cyan-200/40 hover:text-slate-50';

const primaryButtonClassName =
  'inline-flex items-center gap-2 rounded-xl bg-cyan-300 px-4 py-3 text-sm font-semibold text-slate-950 transition hover:bg-cyan-200';

function useDocumentTitle(title: string) {
  useEffect(() => {
    document.title = title;
  }, [title]);
}
