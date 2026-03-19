import { useState } from 'react';
import { Cpu, TrendingUp, Settings, Monitor } from 'lucide-react';
import { Button } from './ui/button';
import { Label } from './ui/label';
import { Input } from './ui/input';
import { Switch } from './ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from './ui/select';

interface PCConfigFormProps {
  onGenerate: (budget: number, usage: string, mode: string, includePeripherals: boolean) => void;
  isLoading: boolean;
}

export function PCConfigForm({ onGenerate, isLoading }: PCConfigFormProps) {
  const [budget, setBudget] = useState('6000');
  const [usage, setUsage] = useState('gaming');
  const [mode, setMode] = useState('balanced');
  const [includePeripherals, setIncludePeripherals] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onGenerate(Number(budget), usage, mode, includePeripherals);
  };

  return (
    <form onSubmit={handleSubmit} className="space-y-5 sm:space-y-6">
      {/* Budget Input */}
      <div className="space-y-2">
        <Label htmlFor="budget" className="flex items-center gap-2 text-slate-100">
          <TrendingUp className="size-4" />
          预算 (¥)
        </Label>
        <Input
          id="budget"
          type="number"
          value={budget}
          onChange={(e) => setBudget(e.target.value)}
          placeholder="输入您的预算"
          min="1000"
          step="100"
          required
          className="h-12 text-base sm:text-lg"
        />
      </div>

      {/* Include Peripherals Switch */}
      <div className="flex flex-col gap-4 rounded-xl border border-cyan-300/24 bg-cyan-400/5 p-4 shadow-[inset_0_0_0_1px_rgba(130,220,255,0.05)] sm:flex-row sm:items-center sm:justify-between">
        <div className="space-y-1">
          <Label htmlFor="peripherals" className="flex items-center gap-2 cursor-pointer text-slate-100">
            <Monitor className="size-4" />
            包含外设
          </Label>
          <p className="pr-2 text-sm leading-6 text-slate-400">
            {includePeripherals ? '预算包含显示器、键盘、鼠标等' : '仅计算主机配置'}
          </p>
        </div>
        <Switch
          id="peripherals"
          checked={includePeripherals}
          onCheckedChange={setIncludePeripherals}
          className="self-end sm:self-auto"
        />
      </div>

      {/* Usage Select */}
      <div className="space-y-2">
        <Label htmlFor="usage" className="flex items-center gap-2 text-slate-100">
          <Cpu className="size-4" />
          用途
        </Label>
        <Select value={usage} onValueChange={setUsage}>
          <SelectTrigger id="usage" className="h-12">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="gaming">游戏</SelectItem>
            <SelectItem value="office">办公</SelectItem>
            <SelectItem value="design">设计/视频剪辑</SelectItem>
            <SelectItem value="programming">编程开发</SelectItem>
            <SelectItem value="ai">AI/深度学习</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Mode Select */}
      <div className="space-y-2">
        <Label htmlFor="mode" className="flex items-center gap-2 text-slate-100">
          <Settings className="size-4" />
          模式
        </Label>
        <Select value={mode} onValueChange={setMode}>
          <SelectTrigger id="mode" className="h-12">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="balanced">均衡</SelectItem>
            <SelectItem value="performance">性能优先</SelectItem>
            <SelectItem value="budget">性价比</SelectItem>
            <SelectItem value="quiet">静音</SelectItem>
          </SelectContent>
        </Select>
      </div>

      {/* Submit Button */}
      <Button 
        type="submit" 
        className="h-12 w-full text-base sm:text-lg"
        disabled={isLoading}
      >
        {isLoading ? '生成中...' : '生成配置方案'}
      </Button>
    </form>
  );
}
