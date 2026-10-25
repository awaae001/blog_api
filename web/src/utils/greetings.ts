import type { Component } from 'vue'
import { Moon, MoonNight, Sunny, Sunrise, Sunset } from '@element-plus/icons-vue'

export interface GreetingPeriod {
  /** 起始小时（含） */
  from: number
  /** 结束小时（不含） */
  to: number
  period: string
  icon: Component
  lines: string[]
}

// 语料池：0-4 午夜 / 5-10 上午 / 11-16 下午 / 17-18 黄昏 / 19-23 夜间
export const greetingPeriods: GreetingPeriod[] = [
  {
    from: 0,
    to: 5,
    period: '午夜',
    icon: MoonNight,
    lines: [
      '您也没有睡啊。',
      '这个点还在忙，辛苦了。',
      '夜深了，写完这段就去睡吧。',
      '整个世界都睡了，只剩你和服务器还醒着。',
      '熬夜写的代码，明早记得再看一遍。',
      '喝口水吧，别只喝咖啡。'
    ]
  },
  {
    from: 5,
    to: 11,
    period: '上午',
    icon: Sunrise,
    lines: [
      '不如抬头看看外面天气如何？',
      '今天外面下雨了吗？',
      '清醒了一会儿，思路顺畅吗？',
      '早饭吃了吗？',
      '新的一天，先看看昨晚有没有友链挂掉。',
      '趁脑子最清楚，把最难的事先做了吧。'
    ]
  },
  {
    from: 11,
    to: 17,
    period: '下午',
    icon: Sunny,
    lines: [
      '今天外面下雨了吗？',
      '下午容易困，要不要伸个懒腰？',
      '外面太阳大吗？',
      '午饭吃得怎么样？',
      '来杯茶，慢慢写。',
      '要是卡住了，起来走两步再回来。'
    ]
  },
  {
    from: 17,
    to: 19,
    period: '黄昏',
    icon: Sunset,
    lines: [
      '窗外的天色应该挺好看了。',
      '趁着这点余光，还能写点东西。',
      '今天的天色怎么样？',
      '快到饭点了，晚饭想吃什么？',
      '今天写了多少，心里有数吗？'
    ]
  },
  {
    from: 19,
    to: 24,
    period: '夜间',
    icon: Moon,
    lines: [
      '夜里安静，正是写东西的好时候。',
      '今天过得怎么样？',
      '再晚一点就该休息了。',
      '要不要翻翻 RSS，看看朋友们最近写了什么？',
      '今天就收个尾吧，明天再继续。',
      '泡个热水澡，然后早点睡。'
    ]
  }
]

export const findGreetingPeriod = (hour: number): GreetingPeriod =>
  greetingPeriods.find((p) => hour >= p.from && hour < p.to) ?? greetingPeriods[0]

export const pickGreetingLine = (period: GreetingPeriod): string =>
  period.lines[Math.floor(Math.random() * period.lines.length)]
