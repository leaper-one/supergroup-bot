import * as echart from 'echarts'

let newUserList: any = []
let totalUserList: any = []

let newMessageList: any = []
let totalMessageList: any = []

let totalUserCount = 0
let totalMessageCount = 0


export function getStatisticsDate(res: any) {
  const { list, today } = res
  initData()
  handleListData(list, today)
  let userObj = {
    name: '用户',
    total: totalUserCount,
    today: today.users,
    data: [
      {
        name: '新用户',
        color: '#1BACC0',
        list: newUserList
      },
      {
        name: '总用户',
        color: '#5099DD',
        list: totalUserList
      }
    ]
  }
  let messageObj = {
    name: '消息',
    total: totalMessageCount,
    today: today.messages,
    data: [
      {
        name: '今日消息',
        color: '#1BACC0',
        list: newMessageList
      },
      {
        name: '总消息',
        color: '#5099DD',
        list: totalMessageList
      }
    ]
  }
  return [userObj, messageObj]
}

function initData() {
  newUserList = []
  totalUserList = []
  newMessageList = []
  totalMessageList = []
  totalUserCount = 0
  totalMessageCount = 0
}

function handleListData(list: any, today: any) {
  if (list.length === 0) return handleDaily(today)
  let todayDate = today.date
  let currentDate = list[0].date
  const mapList: any = transferListToMap(list)
  while (true) {
    if (currentDate === todayDate) break
    handleDaily(mapList[currentDate] || { date: currentDate })
    currentDate = getNextDate(currentDate)
  }
  handleDaily(today)
}

function getNextDate(date: any) {
  date = new Date(Number(new Date(date)) + 86400000)
  return date.toISOString().slice(0, 10)
}

function transferListToMap(list: any) {
  let obj: any = {}
  list.forEach((item: any) => obj[item.date] = item)
  return obj
}

function handleDaily(dataInfo: any) {
  const { date, users = 0, messages = 0 } = dataInfo
  totalUserList.push([date, totalUserCount += users])
  newUserList.push([date, users])
  totalMessageList.push([date, totalMessageCount += messages])
  newMessageList.push([date, messages])
}

export function getOptions(res: any) {
  const { name, total, today, data, options = {}, formatter = null } = res
  let legend: any = [], series: any = [], color: any = []
  data.forEach((_res: any) => {
    const { name, list, color: _color } = _res
    color.push(_color)
    legend.push({ name })
    series.push(getSeries(name, list, _color))
  })
  let addX = String(total).length * 9 + 30
  return {
    color,
    tooltip: {
      trigger: 'axis',
      formatter,
      axisPointer: {
        type: 'line',
        lineStyle: {
          type: 'dashed',
          color: 'red'
        }
      }
    },
    legend: {
      data: legend,
      icon: 'circle',
      type: '',
      itemWidth: 8,
      itemHeight: 8,
      itemGap: 10,
      bottom: 16,
      align: 'auto',
      textStyle: {
        fontSize: 12,
        color: '#A5A7C8',
        padding: [3, 10, 0, 0]
      },
      ...options
    },
    textStyle: {
      color: '#A5A7C8',
      fontFamily: 'Nunito',
      fontSize: 10,
    },
    title: {
      text: name,
      left: 16,
      top: 16,
      textStyle: {
        color: '#4C4471',
        fontWeight: '400',
        fontSize: 18
      },
      subtext: total,
      subtextStyle: {
        color: '#4C4471',
        fontSize: 14,
        lineHeight: 20
      }
    },
    graphic: {
      elements: [
        {
          type: 'text',
          style: {
            text: '+' + today,
            x: addX,
            y: 52,
            font: '12px',
            fill: '#1BACC0'
          }
        }
      ]
    },
    grid: {
      top: '32%',
      left: '12%',
      right: '8%',
      bottom: "70"
    },
    xAxis: {
      type: 'time',
      minInterval: 3600 * 24 * 1000,
      axisLabel: {
        fontSize: 10,
        color: '#A5A7C8',
        formatter: (value: any) => {
          let day = new Date(value).getDate()
          return day < 10 ? '0' + day : day.toString()
        }
      },
      axisLine: {
        show: false
      },
      axisTick: {
        show: false
      },
      splitLine: {
        show: false
      }
    },
    yAxis: {
      type: 'value',
      minInterval: 1,
      axisLabel: {
        fontSize: 10,
        color: '#A5A7C8'
      },
      axisLine: {
        show: false
      },
      axisTick: {
        show: false
      },
      splitLine: {
        lineStyle: {
          color: '#ECEFFF'
        }
      },
      splitNumber: 3,
    },
    series,
    animationDuration: 500,
  }
}


function getSeries(name: string, list: any, color: any) {
  return {
    data: list,
    name: name,
    type: 'line',
    smooth: true,
    showSymbol: false,
    symbol: 'circle',
    symbolSize: 1,
    zlevel: 2,
    hoverAnimation: true,
    itemStyle: {
      borderColor: '#fff',
      borderWidth: 2,
    },
    lineStyle: {
      color
    },
    areaStyle: {
      color: new echart.graphic.LinearGradient(0, 0, 0, 1, [
        { offset: 0, color },
        { offset: 1, color: '#fff' }
      ])
    }
  }
}
