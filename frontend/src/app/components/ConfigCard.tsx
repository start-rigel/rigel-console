import { Check } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from './ui/card';
import { Badge } from './ui/badge';
import { Separator } from './ui/separator';

export interface PCComponent {
  name: string;
  model: string;
  price: number;
}

export interface PCConfig {
  cpu: PCComponent;
  gpu: PCComponent;
  motherboard: PCComponent;
  ram: PCComponent;
  storage: PCComponent;
  psu: PCComponent;
  case: PCComponent;
  cooler: PCComponent;
}

export interface Peripherals {
  monitor: PCComponent;
  keyboard: PCComponent;
  mouse: PCComponent;
}

export type UsageType = 'gaming' | 'office' | 'design' | 'programming' | 'ai';
export type ModeType = 'balanced' | 'performance' | 'budget' | 'quiet';

interface ConfigCardProps {
  config: PCConfig;
  totalPrice: number;
  budget: number;
  usage: UsageType;
  mode: ModeType;
  peripherals?: Peripherals;
  peripheralsPrice?: number;
}

function getUsageLabel(usage: UsageType) {
  const labels: Record<UsageType, string> = {
    gaming: '游戏',
    office: '办公',
    design: '设计 / 剪辑',
    programming: '编程开发',
    ai: 'AI / 深度学习',
  };

  return labels[usage];
}

function getModeLabel(mode: ModeType) {
  const labels: Record<ModeType, string> = {
    balanced: '均衡',
    performance: '性能优先',
    budget: '性价比',
    quiet: '静音',
  };

  return labels[mode];
}

function getSummary(config: PCConfig, usage: UsageType, mode: ModeType, isWithinBudget: boolean) {
  const summaryByUsage: Record<UsageType, string> = {
    gaming: `这套配置把预算重点放在 ${config.gpu.name} 和 ${config.cpu.name} 上，优先保证游戏帧率稳定，不把钱浪费在感知不强的堆料上。`,
    office: `这套配置优先控制整机成本和稳定性，保留日常办公、网页多开和轻量多任务的流畅体验。`,
    design: `这套配置在处理器、多内存和存储速度之间做了平衡，更适合图形设计、剪辑导出和多软件并行。`,
    programming: `这套配置优先照顾编译、容器和多任务并行的体验，避免把预算压在对开发收益不高的部件上。`,
    ai: `这套配置把重点放在显卡显存和多线程处理能力上，适合本地推理、模型实验和较高负载场景。`,
  };

  const modeByLabel: Record<ModeType, string> = {
    balanced: '整体取向偏稳，不会为了单点参数牺牲其他关键部件。',
    performance: '这版更偏向直接性能，预算会优先向核心算力部件倾斜。',
    budget: '这版强调有效投入，优先保住体感收益最大的部分。',
    quiet: '这版会尽量避免高噪声组合，适合长时间近距离使用。',
  };

  return `${summaryByUsage[usage]}${modeByLabel[mode]}${isWithinBudget ? ' 当前总价仍控制在你的预算内。' : ' 当前方案性能更积极，但价格已经超过预算。'}`;
}

function getRecommendationNotes(config: PCConfig, usage: UsageType, mode: ModeType) {
  return [
    {
      title: '这套方案优先保证什么',
      detail: usage === 'gaming'
        ? `优先保证 ${config.gpu.name} 的图形性能，其次用 ${config.cpu.name} 避免高帧率场景下的明显瓶颈。`
        : usage === 'ai'
          ? `优先保证显卡算力和多线程处理能力，让训练、推理和并行任务更稳定。`
          : `优先保证 ${config.cpu.name} 与 ${config.ram.name} 的协同表现，兼顾流畅度和长期可用性。`,
    },
    {
      title: '预算主要花在了哪里',
      detail: `当前主机预算的主要支出集中在 ${config.gpu.name}、${config.cpu.name} 和 ${config.storage.name}，这是最影响体感的三个部分。`,
    },
    {
      title: '如果你还想继续优化',
      detail: mode === 'budget'
        ? '建议下一步先升级 SSD 容量或散热器，而不是盲目抬高主板等级。'
        : mode === 'quiet'
          ? '若还想更安静，可以继续换更高等级风冷或更低转速风扇方案。'
          : '如果后续预算增加，优先升级显卡或增加存储容量，收益通常最明显。',
    },
  ];
}

export function ConfigCard({ config, totalPrice, budget, usage, mode, peripherals, peripheralsPrice }: ConfigCardProps) {
  const components = [
    { label: 'CPU', data: config.cpu },
    { label: 'GPU', data: config.gpu },
    { label: '主板', data: config.motherboard },
    { label: '内存', data: config.ram },
    { label: '存储', data: config.storage },
    { label: '电源', data: config.psu },
    { label: '机箱', data: config.case },
    { label: '散热器', data: config.cooler },
  ];

  const peripheralComponents = peripherals ? [
    { label: '显示器', data: peripherals.monitor },
    { label: '键盘', data: peripherals.keyboard },
    { label: '鼠标', data: peripherals.mouse },
  ] : [];

  const grandTotal = totalPrice + (peripheralsPrice || 0);
  const isWithinBudget = grandTotal <= budget;
  const budgetDiff = Math.abs(grandTotal - budget);
  const budgetPercentage = ((grandTotal / budget) * 100).toFixed(1);
  const summary = getSummary(config, usage, mode, isWithinBudget);
  const recommendationNotes = getRecommendationNotes(config, usage, mode);

  return (
    <Card className="overflow-hidden border-cyan-300/26 bg-slate-950/88 shadow-[0_18px_60px_rgba(0,0,0,0.3),inset_0_0_0_1px_rgba(130,220,255,0.08)]">
      <CardHeader className="border-b border-cyan-300/20 bg-slate-950/98">
        <div className="flex flex-col items-start gap-3 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <CardTitle className="text-xl text-slate-50 sm:text-2xl">推荐配置方案</CardTitle>
            <CardDescription className="mt-1 text-slate-400">
              基于您的预算和需求智能匹配
            </CardDescription>
            <div className="mt-4 flex flex-wrap gap-2">
              <Badge variant="secondary">{getUsageLabel(usage)}</Badge>
              <Badge variant="secondary">{getModeLabel(mode)}</Badge>
              <Badge variant="secondary">{peripherals ? '包含外设' : '仅主机'}</Badge>
            </div>
          </div>
          <Badge 
            variant={isWithinBudget ? "default" : "destructive"}
            className="text-sm"
          >
            {isWithinBudget ? '符合预算' : '超出预算'}
          </Badge>
        </div>
        
        <div className="mt-4 flex flex-col items-start gap-3 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-sm text-muted-foreground">总价</p>
            <p className="text-2xl font-bold sm:text-3xl">¥{grandTotal.toLocaleString()}</p>
          </div>
          <div className="text-left sm:text-right">
            <p className="text-sm text-muted-foreground">预算利用率</p>
            <p className="text-xl font-semibold">{budgetPercentage}%</p>
          </div>
        </div>

        {!isWithinBudget && (
          <p className="mt-2 text-sm text-red-600 dark:text-red-400">
            超出预算 ¥{budgetDiff.toLocaleString()}
          </p>
        )}
      </CardHeader>

      <CardContent className="pt-6">
        <div className="rounded-xl border border-cyan-300/24 bg-cyan-400/4 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
          <p className="text-xs font-semibold uppercase tracking-[0.18em] text-cyan-300/72">
            顾问判断
          </p>
          <p className="mt-2 text-sm leading-7 text-slate-200">
            {summary}
          </p>
        </div>

        <div className="mt-5 grid gap-3 md:grid-cols-3">
          {recommendationNotes.map((note) => (
            <div key={note.title} className="rounded-xl border border-cyan-300/24 bg-slate-900/82 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
              <p className="text-sm font-semibold text-slate-100">{note.title}</p>
              <p className="mt-2 text-sm leading-6 text-slate-400">{note.detail}</p>
            </div>
          ))}
        </div>

        <Separator className="my-6" />

        {/* Main PC Components */}
        <div>
          <h3 className="mb-4 flex items-center gap-2 font-semibold text-slate-100">
            <span className="text-blue-600 dark:text-blue-400">●</span>
            主机配置
          </h3>
          <div className="space-y-4">
            {components.map((component, index) => (
              <div key={index} className="rounded-xl border border-cyan-300/20 bg-slate-900/72 px-4 py-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]">
                <div className="flex flex-col items-start gap-3 sm:flex-row sm:justify-between sm:gap-4">
                  <div className="flex-1 min-w-0">
                    <div className="flex items-center gap-2">
                      <Check className="size-4 shrink-0 text-cyan-300" />
                      <p className="text-sm font-medium text-slate-100">{component.label}</p>
                    </div>
                    <p className="ml-6 mt-1 break-words text-sm text-slate-300">
                      {component.data.name}
                    </p>
                    <p className="ml-6 mt-0.5 break-words text-xs leading-5 text-slate-500">
                      {component.data.model}
                    </p>
                  </div>
                  <div className="shrink-0 pl-6 text-left sm:pl-0 sm:text-right">
                    <p className="font-semibold text-cyan-200">
                      ¥{component.data.price.toLocaleString()}
                    </p>
                  </div>
                </div>
                {index < components.length - 1 && <Separator className="mt-4 bg-cyan-300/12" />}
              </div>
            ))}
          </div>

          <div className="mt-4 flex items-center justify-between rounded-xl border border-cyan-300/24 bg-slate-900/78 px-4 py-4">
            <span className="font-medium text-slate-200">主机小计</span>
            <span className="text-lg font-bold text-cyan-200">
              ¥{totalPrice.toLocaleString()}
            </span>
          </div>
        </div>

        {/* Peripherals Section */}
        {peripherals && peripheralsPrice && (
          <>
            <Separator className="my-6" />
            <div>
              <h3 className="mb-4 flex items-center gap-2 font-semibold text-slate-100">
                <span className="text-purple-600 dark:text-purple-400">●</span>
                外设配置
              </h3>
              <div className="space-y-4">
                {peripheralComponents.map((component, index) => (
                  <div key={index} className="rounded-xl border border-cyan-300/20 bg-slate-900/72 px-4 py-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.04)]">
                <div className="flex flex-col items-start gap-3 sm:flex-row sm:justify-between sm:gap-4">
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center gap-2">
                          <Check className="size-4 shrink-0 text-cyan-300" />
                          <p className="text-sm font-medium text-slate-100">{component.label}</p>
                        </div>
                        <p className="ml-6 mt-1 break-words text-sm text-slate-300">
                          {component.data.name}
                        </p>
                        <p className="ml-6 mt-0.5 break-words text-xs leading-5 text-slate-500">
                          {component.data.model}
                        </p>
                      </div>
                      <div className="shrink-0 pl-6 text-left sm:pl-0 sm:text-right">
                        <p className="font-semibold text-cyan-200">
                          ¥{component.data.price.toLocaleString()}
                        </p>
                      </div>
                    </div>
                    {index < peripheralComponents.length - 1 && <Separator className="mt-4 bg-cyan-300/12" />}
                  </div>
                ))}
              </div>

              <div className="mt-4 flex items-center justify-between rounded-xl border border-cyan-300/24 bg-slate-900/78 px-4 py-4">
                <span className="font-medium text-slate-200">外设小计</span>
                <span className="text-lg font-bold text-cyan-200">
                  ¥{peripheralsPrice.toLocaleString()}
                </span>
              </div>
            </div>
          </>
        )}

        <Separator className="my-6 bg-cyan-300/14" />

        <div className="flex items-center justify-between rounded-2xl border border-cyan-300/28 bg-cyan-400/6 px-4 py-4 text-lg shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)]">
          <span className="font-semibold text-slate-100">总计</span>
          <span className="text-2xl font-bold text-cyan-200">
            ¥{grandTotal.toLocaleString()}
          </span>
        </div>
      </CardContent>
    </Card>
  );
}
