import { getI18n } from "@/locales/tools"
import { $set } from "@/stores/localStorage"

$set("umi_locale", navigator.language.includes("zh") ? "zh-CN" : "en-US")

const _i18n = {
  site: {
    title: "超级社群"
  },
  pre: {
    create: {
      title: "开通社群",
      desc:
        "社群助手可以自动创建新群、定时检测持仓、群发红包、发布公告等功能。",
      button: "支付 0.01 XIN 开通",
      action: "创建社群",
    },
    explore: {
      title: "社群助手",
      desc: "请把社群助手添加到群组再设置为管理员，然后开启社群管理功能。",
      button: "发现社群",
    },
  },

  setting: {
    title: "设置",
    accept: "接收聊天信息",
    acceptTips: "停止接收社群聊天仍然可以收到公告，所有重要的项目动态都会发公告！",
    newNotice: "新人入群提醒",
    receivedFirst: "请先接收聊天信息",
    auth: "重新授权",
    authConfirm: "确认重新授权吗？",
    exit: "退出社群",
    exitConfirm: "确认退出社群吗？",
    cancel: {
      title: "停止接受聊天",
      content: "停止接收社群聊天仍然可以收到公告，所有重要的项目动态都会发公告！<br /> 依次输入下方数字确认操作。"
    },
    exited: "社群已退出",
    exitedDesc: "你已成功退出社群，你账户相关的数据均已删除，点右上角关闭社群页面即可，欢迎再来！"
  },

  manager: {
    setting: "设置",
    base: "基本设置",
    description: "社群简介",
    welcome: "入群欢迎语",
    high: "高级管理",
  },
  broadcast: {
    title: "公告管理",
    holder: "请填写公告",
    recall: "撤回",
    confirmRecall: "确认撤回吗",
    recallSuccess: "撤回成功",
    status0: "发送中",
    status1: "已发布",
    status2: "已撤回",
    checkNumber: "请检查数字是否一致",
    sent: "群发公告",
    input: "依次输入上方数字群发公告",
    fill: "请先填写公告",
    send: "发送",
  },
  stat: {
    title: "数据统计",
    totalUser: "用户总量",
    highUser: "最低持仓用户",
    weekUser: "周新增用户",
    weekActiveUser: "周活跃用户",
    totalMessage: "消息总数",
    weekMessage: "周发消息",
    all: "全部",
    month: "最近一月",
    week: "最近一周",
    user: "用户",
    newUser: "新用户",
    activeUser: "活跃用户",
    msg: "消息",
    dailyMsg: "日消息",
    totalMsg: "总消息"
  },
  member: {
    title: "成员管理",
    status8: "嘉宾",
    status9: "管理员",
    hour: "最近 {n} 小时活跃",
    day: "最近 {n} 天活跃",
    month: "最近 {n} 月活跃",
    year: "最近 {n} 年活跃",
    action: {
      set: "设为{c}",
      cancel: "取消{c}",
      confirmSet: "确认将 {full_name}({identity_number}) 设为{c} 吗？",
      confirmCancel: "确认取消 {full_name}({identity_number}) 为{c} 吗？",
      guest: "嘉宾",
      admin: "管理员",
      mute: "禁言",
      confirmMute: "确认禁言 {full_name}({identity_number}) {mute_time} 小时吗？",
      block: "拉黑",
      confirmBlock: "确认拉黑 {full_name}({identity_number}) 吗？",
    },
    modal: {
      unit: "小时",
      desc: "用户将被禁言 1 小时，禁言不影响用户接受消息和抢红包。"
    },
    status: {
      title: "成员类型",
      all: "全部",
      guest: "嘉宾",
      admin: "管理员",
      mute: "禁言",
      block: "拉黑",
      people: "人"
    },
    done: "到底了",
  },
  join: {
    title: "发现社群",
    received: "领取成功",

    main: {
      join: "授权入群",
      joinTips: "【风险提示】 Mixin 不为任何项目做价格背书、项目担保。",

      appointBtn: "预约",
      appointedBtn: "已预约",
      appointedTips: "添加联系人",

      receiveBtn: "领取空投",

      receivedBtn: "已领取空投",
      receivedTips: "加入空投群",

      noAccess: "没有资格",

      appointOver: "空投已结束",
    },

    modal: {
      auth: "授权检测不通过",
      authDesc: "请同意授权查询您的资产，数据仅用于持仓检测。",
      authBtn: "再次授权",

      forbid: "禁止加入群组",
      forbidDesc1: "24 小时内不能加入群组，请联系管理员或等 24 小时再进入群。",
      forbidDesc2: "你被禁止入群，想要加入群组请联系管理员。",
      forbidBtn: "知道了",

      shares: "持仓检测不通过",
      sharesBtn: "再次检测",
      sharesFail: "检测不通过",
      sharesTips: "立刻购买",
      sharesCheck: "请确保您的资产满足一下任意持仓：",
      sharesCheck1: "不小于",
      sharesCheck2:
        "持仓检测支持 Mixin 钱包、 Exin 流动池、活期宝、省心投和 Fox 活期理财、定期理财、可盈池。",

      appoint: "预约成功",
      appointDesc:
        "感谢预约！请添加当前机器人为联系人并打开通知权限，以便及时收到空投领取资格通知。",
      appointBtn: "添加联系人",
      appointTips: "点右上角关闭机器人等待通知即可",
      receive: "空投奖励",
      receiveDesc:
        "恭喜获得 MobileCoin 空投资格！感谢你对 MobileCoin 的支持，欢迎领取空投并入群。",
      receiveBtn: "领取 MOB 空投",
      receivedDesc:
        "这是您参与 {comment} 的空投奖励！ 感谢您对 MobileCoin 和 Mixin 的鼎力支持！",
      receivedBtn: "已领取 MOB 空投",
    },

    code: {
      invite: "用 Mixin Messenger 扫码{action}",
      download: "下载 Mixin Messenger",
      downloadXinsheng: "下载新生大讲堂",
      action: {
        appoint: "预约",
        join: "入群",
      },
    },

    search: {
      name: "群名",
      holder: "发言要求",
      or: "或",
      people: "人",
    },
  },

  home: {
    title: "社群助手",

    people_count: "社群人数",
    week: "本周",

    trade: "交易",
    invite: "邀请入群",
    findGroup: "发现社群",
    findBot: "发现机器人",
    activity: "活动",
    redPacket: "发红包",
    article: "资讯",
    more: "更多活动",
  },

  news: {
    all: "全部",
    replay: "回放",
    broadcast: "公告",
    sendBroadcast: "发公告",
    sendLive: "添加直播预告",
    live: "直播",
    confirmStart: "确认开始直播吗？",
    confirmEnd: "确认结束直播吗？",
    prompt: "请输入直播的地址。",
    form: {
      img: "直播图",
      category: "直播类型",

      "1": "视频直播",
      "2": "图片+语音直播",

      user: "直播嘉宾",
      title: "直播标题",
      desc: "直播简介",
    },
    livePreview: "直播预告",
    action: {
      stop: "停止直播",
      delete: "删除",
      edit: "编辑预告",
      share: "分享预告",
      start: "开始直播",
      top: "置顶",
      cancelTop: "取消置顶"
    },
    confirmTop: "确认置顶吗？",
    confirmCancelTop: "确认取消置顶吗？",

    liveReplay: {
      title: "直播回放",
      delete: "删除"
    },
    stat: {
      title: "直播数据",
      read_count: "观看用户",
      deliver_count: "广播用户",
      duration: "直播时长（分钟）",
      user_count: "发言人数（视频）",
      msg_count: "发言数量（视频）"
    }
  },

  invite: {
    title: "邀请入群",
    desc: "入群邀请",
    card: "发送邀请卡片",
    link: "复制入群链接",
    tip1: "管理员也可以通过成员管理直接添加成员",
    tipNotOpen: "当前社群还没有开启邀请奖励",
    tipOpen:
      "邀请入群奖励已开启，邀请奖励红包请在 48 小时内领取，否则将过期无法领取。<br /><br />" +
      "请务必将邀请卡片或入群链接直接发给你的 Mixin 联系人！！！只有被邀请人在与你的单人会话中打开邀请卡片或邀请链接入群才算有效邀请，️通过群、机器人、浏览器等入群均不计入你的邀请！<br /><br />" +
      "请不要骚扰陌生人，一旦检测到当前用户被举报过多，立刻取消邀请奖励资格。",

    my: {
      title: "我的邀请",
      reward: "邀请奖励",
      people: "邀请人数",
      "0": "等待生效",
      "1": "有效邀请",
      "2": "有效邀请",

      noTitle: "未开启",
      noTips: "后续开启邀请奖励,之前的邀请仍然有效",

      noInvited: "没有邀请",
      rule: "查看规则",
    },
  },

  transfer: {
    title: "交易 {name}",
    price: "价格",
    pool: "资金池",
    earn: "24H 做市年化",
    amount: "24H 交易量",
    method: "交易方式",
    noPrice: "暂无价格",

    order: "最大下单 {amount} {symbol}",

    maker: "自动做市商",
    taker: "{exchange}代购",

    Huobi: "火币",
    BigONE: "BigONE",
    Binance: "币安",
    ExinSwap: "ExinSwap",
    MixSwap: "MixSwap",
    exchange: "交易所",
    sign: "个人多签交易",

    coin: "币币交易",
    otc: "法币入金",

    auth: "蓝盾认证",
    identity: "已实名",

    payMethod: "支付方式",
    bank: "银行",
    alipay: "支付宝",
    wechatpay: "微信",

    category: "币种",
    limit: "额度",
    in5minRate: "5 分钟完成率",
    orderSuccessRank: "订单完成率",
    multisigOrderCount: "总交易数",
  },

  reward: {
    title: "打赏",
    who: "给谁打赏？",
    amount: "数量",
    success: "打赏成功",
  },


  //
  // manager: {
  //   members: "用户总量",
  //   broadcasts: "公告次数",
  //   conversations: "小群数量",
  //   list: "新增用户",
  //
  //   asset: {
  //     title: "资产中心",
  //     total: "总资产",
  //     deposit: "充值",
  //     withdrawal: "提现",
  //     packet_send: "发红包",
  //     packet_refund: "红包返还",
  //     airdrop: "空投奖励",
  //     exin_otc: "社群返佣",
  //
  //     action: {
  //       deposit: "支付",
  //       withdrawal: "提现",
  //     },
  //     checking: "正在检查支付状态",
  //     depositSuccess: "支付成功",
  //     withdrawalSuccess: "提现成功",
  //   },
  // },

  modal: {
    check: "正在检查支付结果",
    loading: "正在加载",
  },

  action: {
    tips: "提示",
    cancel: "取消",
    save: "保存",
    confirm: "确认",
    submit: "提交",
    continue: "继续"
  },

  success: {
    copy: "复制成功",
    send: "发送成功",
    operator: "操作成功",
    save: "保存成功"
  },
  error: {
    people: "人数有误",
    amount: "金额有误",
    mixin: "请在 Mixin 客户端内打开",
    empty: "不能为空",
  },
}
const i18n = {}
getI18n(_i18n, i18n)
export default i18n
