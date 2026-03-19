import { PCConfig, Peripherals } from '../components/ConfigCard';

// Mock data generator for PC configurations
export function generatePCConfig(budget: number, usage: string, mode: string, includePeripherals: boolean = false): PCConfig {
  // If peripherals are included, reserve budget for them
  let pcBudget = budget;
  if (includePeripherals) {
    // Reserve 15-25% of budget for peripherals depending on total budget
    const peripheralPercentage = budget < 5000 ? 0.20 : budget < 10000 ? 0.18 : 0.15;
    pcBudget = Math.floor(budget * (1 - peripheralPercentage));
  }

  // Different configurations based on usage and budget
  const configs: Record<string, (budget: number, mode: string) => PCConfig> = {
    gaming: generateGamingConfig,
    office: generateOfficeConfig,
    design: generateDesignConfig,
    programming: generateProgrammingConfig,
    ai: generateAIConfig,
  };

  const generator = configs[usage] || generateGamingConfig;
  return generator(pcBudget, mode);
}

export function generatePeripherals(budget: number, usage: string): Peripherals {
  const peripheralBudget = budget < 5000 ? 0.20 : budget < 10000 ? 0.18 : 0.15;
  const totalPeripheralBudget = Math.floor(budget * peripheralBudget);

  // Allocate budget: 60% monitor, 25% keyboard, 15% mouse
  const monitorBudget = Math.floor(totalPeripheralBudget * 0.6);
  const keyboardBudget = Math.floor(totalPeripheralBudget * 0.25);
  const mouseBudget = Math.floor(totalPeripheralBudget * 0.15);

  return {
    monitor: getMonitorByBudget(monitorBudget, usage),
    keyboard: getKeyboardByBudget(keyboardBudget, usage),
    mouse: getMouseByBudget(mouseBudget, usage),
  };
}

function getMonitorByBudget(budget: number, usage: string): { name: string; model: string; price: number } {
  if (usage === 'gaming') {
    if (budget >= 2000) {
      return { name: 'LG 27GP850', model: '27寸 2K 180Hz IPS', price: 2299 };
    } else if (budget >= 1200) {
      return { name: '小米 27寸显示器', model: '27寸 2K 165Hz Fast IPS', price: 1399 };
    } else {
      return { name: 'AOC 24G2', model: '24寸 1080P 144Hz IPS', price: 899 };
    }
  } else if (usage === 'design') {
    if (budget >= 3000) {
      return { name: 'Dell U2723DE', model: '27寸 2K IPS 色准显示器', price: 3199 };
    } else if (budget >= 1500) {
      return { name: 'BenQ PD2705U', model: '27寸 4K IPS 专业显示器', price: 2499 };
    } else {
      return { name: 'LG 27UP550', model: '27寸 4K IPS', price: 1699 };
    }
  } else {
    if (budget >= 1500) {
      return { name: 'Dell U2723DE', model: '27寸 2K IPS USB-C', price: 1899 };
    } else if (budget >= 800) {
      return { name: 'LG 24寸显示器', model: '24寸 1080P IPS', price: 799 };
    } else {
      return { name: 'AOC 24B2XH', model: '24寸 1080P IPS', price: 599 };
    }
  }
}

function getKeyboardByBudget(budget: number, usage: string): { name: string; model: string; price: number } {
  if (usage === 'gaming') {
    if (budget >= 600) {
      return { name: '海盗船 K70 RGB', model: '机械键盘 Cherry轴', price: 699 };
    } else if (budget >= 300) {
      return { name: '达尔优 A87', model: '机械键盘 凯华轴', price: 349 };
    } else {
      return { name: '雷柏 V500PRO', model: '机械键盘 雷柏轴', price: 199 };
    }
  } else if (usage === 'programming' || usage === 'design') {
    if (budget >= 800) {
      return { name: 'HHKB Professional', model: '静电容键盘', price: 1699 };
    } else if (budget >= 400) {
      return { name: '阿米洛 VA87M', model: '机械键盘 Cherry轴', price: 599 };
    } else {
      return { name: 'ikbc C87', model: '机械键盘 Cherry轴', price: 399 };
    }
  } else {
    if (budget >= 300) {
      return { name: '罗技 K580', model: '无线键盘', price: 299 };
    } else {
      return { name: '雷柏 MT510', model: '无线键盘', price: 149 };
    }
  }
}

function getMouseByBudget(budget: number, usage: string): { name: string; model: string; price: number } {
  if (usage === 'gaming') {
    if (budget >= 400) {
      return { name: '罗技 G PRO X', model: '无线游戏鼠标 25600DPI', price: 599 };
    } else if (budget >= 200) {
      return { name: '雷蛇 炼狱蝰蛇V3', model: '有线游戏鼠标 30000DPI', price: 299 };
    } else {
      return { name: '达尔优 EM945', model: '无线游戏鼠标', price: 149 };
    }
  } else if (usage === 'design') {
    if (budget >= 500) {
      return { name: '罗技 MX Master 3S', model: '无线办公鼠标', price: 699 };
    } else if (budget >= 200) {
      return { name: '罗技 MX Anywhere 3', model: '无线便携鼠标', price: 399 };
    } else {
      return { name: '雷柏 MT750', model: '无线多模鼠标', price: 199 };
    }
  } else {
    if (budget >= 200) {
      return { name: '罗技 MX Master 3S', model: '无线办公鼠标', price: 399 };
    } else if (budget >= 100) {
      return { name: '罗技 M590', model: '无线静音鼠标', price: 149 };
    } else {
      return { name: '雷柏 M300', model: '无线鼠标', price: 79 };
    }
  }
}

function generateGamingConfig(budget: number, mode: string): PCConfig {
  if (budget >= 10000) {
    return {
      cpu: { name: 'AMD Ryzen 7 7800X3D', model: '8核16线程 4.2GHz', price: 2599 },
      gpu: { name: 'NVIDIA RTX 4070 Ti SUPER', model: '16GB GDDR6X', price: 5799 },
      motherboard: { name: 'MSI B650 TOMAHAWK', model: 'ATX AM5', price: 1399 },
      ram: { name: 'Kingston Fury DDR5', model: '32GB 6000MHz', price: 799 },
      storage: { name: '三星 990 PRO', model: '1TB NVMe SSD', price: 799 },
      psu: { name: '海韵 FOCUS GX-850', model: '850W 80+ Gold', price: 899 },
      case: { name: '联力 O11 Dynamic EVO', model: 'ATX 中塔', price: 899 },
      cooler: { name: '利民 Peerless Assassin 120', model: '双塔散热器', price: 249 },
    };
  } else if (budget >= 7000) {
    return {
      cpu: { name: 'AMD Ryzen 5 7600X', model: '6核12线程 4.7GHz', price: 1499 },
      gpu: { name: 'NVIDIA RTX 4060 Ti', model: '8GB GDDR6', price: 3199 },
      motherboard: { name: 'MSI B650M MORTAR', model: 'M-ATX AM5', price: 999 },
      ram: { name: 'Kingston Fury DDR5', model: '16GB 6000MHz', price: 449 },
      storage: { name: '西数 SN770', model: '1TB NVMe SSD', price: 499 },
      psu: { name: '安钛克 NE650', model: '650W 80+ Bronze', price: 399 },
      case: { name: '先马 趣造I', model: 'M-ATX 中塔', price: 199 },
      cooler: { name: '利民 PA120 SE', model: '单塔散热器', price: 129 },
    };
  } else {
    return {
      cpu: { name: 'Intel i5-12400F', model: '6核12线程 4.4GHz', price: 999 },
      gpu: { name: 'NVIDIA RTX 4060', model: '8GB GDDR6', price: 2399 },
      motherboard: { name: '华硕 B660M-K', model: 'M-ATX LGA1700', price: 599 },
      ram: { name: 'Kingston DDR4', model: '16GB 3200MHz', price: 299 },
      storage: { name: '金士顿 NV2', model: '512GB NVMe SSD', price: 299 },
      psu: { name: '长城 巨龙500W', model: '500W 80+ Bronze', price: 249 },
      case: { name: '爱国者 月光宝盒T20', model: 'M-ATX 中塔', price: 149 },
      cooler: { name: '九州风神 玄冰400', model: '塔式散热器', price: 79 },
    };
  }
}

function generateOfficeConfig(budget: number, mode: string): PCConfig {
  if (budget >= 5000) {
    return {
      cpu: { name: 'Intel i5-13400', model: '10核16线程 4.6GHz', price: 1399 },
      gpu: { name: '核显', model: 'Intel UHD 730', price: 0 },
      motherboard: { name: '华硕 B760M-K', model: 'M-ATX LGA1700', price: 699 },
      ram: { name: 'Kingston DDR5', model: '16GB 4800MHz', price: 399 },
      storage: { name: '三星 980', model: '512GB NVMe SSD', price: 399 },
      psu: { name: '航嘉 WD500K', model: '500W 80+ White', price: 199 },
      case: { name: '金河田 峥嵘Z30', model: 'M-ATX 中塔', price: 99 },
      cooler: { name: 'Intel 原装散热器', model: '下压式', price: 0 },
    };
  } else {
    return {
      cpu: { name: 'Intel i3-12100', model: '4核8线程 4.3GHz', price: 799 },
      gpu: { name: '核显', model: 'Intel UHD 730', price: 0 },
      motherboard: { name: '华硕 H610M-K', model: 'M-ATX LGA1700', price: 499 },
      ram: { name: 'Kingston DDR4', model: '16GB 3200MHz', price: 299 },
      storage: { name: '致钛 TiPlus5000', model: '512GB NVMe SSD', price: 249 },
      psu: { name: '长城 静音大师400W', model: '400W 80+ Bronze', price: 179 },
      case: { name: '先马 工匠5号', model: 'M-ATX 中塔', price: 79 },
      cooler: { name: 'Intel 原装散热器', model: '下压式', price: 0 },
    };
  }
}

function generateDesignConfig(budget: number, mode: string): PCConfig {
  if (budget >= 12000) {
    return {
      cpu: { name: 'AMD Ryzen 9 7950X', model: '16核32线程 4.5GHz', price: 3999 },
      gpu: { name: 'NVIDIA RTX 4070', model: '12GB GDDR6X', price: 4599 },
      motherboard: { name: 'MSI X670E CARBON', model: 'ATX AM5', price: 2399 },
      ram: { name: 'Kingston Fury DDR5', model: '64GB 6000MHz', price: 1799 },
      storage: { name: '三星 990 PRO', model: '2TB NVMe SSD', price: 1599 },
      psu: { name: '海韵 FOCUS GX-1000', model: '1000W 80+ Gold', price: 1199 },
      case: { name: '联力 O11 Dynamic', model: 'ATX 中塔', price: 799 },
      cooler: { name: '九州风神 船长360', model: '360mm 一体水冷', price: 799 },
    };
  } else {
    return {
      cpu: { name: 'AMD Ryzen 7 7700X', model: '8核16线程 4.5GHz', price: 1999 },
      gpu: { name: 'NVIDIA RTX 4060 Ti', model: '16GB GDDR6', price: 3499 },
      motherboard: { name: 'MSI B650M MORTAR', model: 'M-ATX AM5', price: 999 },
      ram: { name: 'Kingston Fury DDR5', model: '32GB 6000MHz', price: 799 },
      storage: { name: '西数 SN770', model: '1TB NVMe SSD', price: 499 },
      psu: { name: '海韵 FOCUS GX-750', model: '750W 80+ Gold', price: 799 },
      case: { name: '先马 趣造II', model: 'M-ATX 中塔', price: 249 },
      cooler: { name: '利民 FC140', model: '双塔散热器', price: 299 },
    };
  }
}

function generateProgrammingConfig(budget: number, mode: string): PCConfig {
  if (budget >= 8000) {
    return {
      cpu: { name: 'AMD Ryzen 7 7700X', model: '8核16线程 4.5GHz', price: 1999 },
      gpu: { name: 'NVIDIA RTX 4060', model: '8GB GDDR6', price: 2399 },
      motherboard: { name: 'MSI B650M MORTAR', model: 'M-ATX AM5', price: 999 },
      ram: { name: 'Kingston Fury DDR5', model: '32GB 6000MHz', price: 799 },
      storage: { name: '三星 990 PRO', model: '1TB NVMe SSD', price: 799 },
      psu: { name: '安钛克 NE650', model: '650W 80+ Bronze', price: 399 },
      case: { name: '先马 趣造I', model: 'M-ATX 中塔', price: 199 },
      cooler: { name: '利民 PA120', model: '单塔散热器', price: 179 },
    };
  } else {
    return {
      cpu: { name: 'AMD Ryzen 5 7600', model: '6核12线程 3.8GHz', price: 1299 },
      gpu: { name: '核显', model: 'AMD Radeon Graphics', price: 0 },
      motherboard: { name: 'MSI B650M-P', model: 'M-ATX AM5', price: 799 },
      ram: { name: 'Kingston Fury DDR5', model: '32GB 5200MHz', price: 649 },
      storage: { name: '西数 SN770', model: '1TB NVMe SSD', price: 499 },
      psu: { name: '长城 巨龙500W', model: '500W 80+ Bronze', price: 249 },
      case: { name: '先马 工匠5号', model: 'M-ATX 中塔', price: 79 },
      cooler: { name: 'AMD 原装散热器', model: '下压式', price: 0 },
    };
  }
}

function generateAIConfig(budget: number, mode: string): PCConfig {
  if (budget >= 15000) {
    return {
      cpu: { name: 'AMD Ryzen 9 7950X', model: '16核32线程 4.5GHz', price: 3999 },
      gpu: { name: 'NVIDIA RTX 4080 SUPER', model: '16GB GDDR6X', price: 7999 },
      motherboard: { name: 'MSI X670E CARBON', model: 'ATX AM5', price: 2399 },
      ram: { name: 'Kingston Fury DDR5', model: '64GB 6000MHz', price: 1799 },
      storage: { name: '三星 990 PRO', model: '2TB NVMe SSD', price: 1599 },
      psu: { name: '海韵 PRIME TX-1000', model: '1000W 80+ Titanium', price: 1999 },
      case: { name: '联力 O11 Dynamic XL', model: 'E-ATX 全塔', price: 1199 },
      cooler: { name: '海盗船 H150i', model: '360mm 一体水冷', price: 999 },
    };
  } else {
    return {
      cpu: { name: 'AMD Ryzen 9 7900X', model: '12核24线程 4.7GHz', price: 2799 },
      gpu: { name: 'NVIDIA RTX 4070 SUPER', model: '12GB GDDR6X', price: 4899 },
      motherboard: { name: 'MSI B650 TOMAHAWK', model: 'ATX AM5', price: 1399 },
      ram: { name: 'Kingston Fury DDR5', model: '32GB 6000MHz', price: 799 },
      storage: { name: '三星 990 PRO', model: '1TB NVMe SSD', price: 799 },
      psu: { name: '海韵 FOCUS GX-850', model: '850W 80+ Gold', price: 899 },
      case: { name: '联力 O11 Dynamic', model: 'ATX 中塔', price: 799 },
      cooler: { name: '九州风神 船长280', model: '280mm 一体水冷', price: 599 },
    };
  }
}

export function calculateTotalPrice(config: PCConfig): number {
  return Object.values(config).reduce((total, component) => total + component.price, 0);
}

export function calculatePeripheralsPrice(peripherals: Peripherals): number {
  return peripherals.monitor.price + peripherals.keyboard.price + peripherals.mouse.price;
}