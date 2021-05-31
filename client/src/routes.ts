export default [
  { path: "/", component: "@/pages/index" },
  { path: "/auth", component: "@/pages/auth" },
  // { path: "/about", component: "@/pages/about" },
  { path: "/pre", component: "@/pages/pre/index" }, // 从机器人打开首页
  { path: "/join/:number", component: "@/pages/pre/join" }, // 申请加入持仓群页面
  { path: "/join", component: "@/pages/pre/join" }, // 申请加入持仓群页面
  { path: "/explore", component: "@/pages/pre/search", title: "发现社群" }, // 探索其他社群页面
  { path: "/pre/search", redirect: "/explore" }, // 兼容处理

  { path: "/home", component: "@/pages/home/index" },
  {
    path: "/invite",
    component: "@/pages/home/invite/index",
    title: "邀请入群",
  },
  // {
  //   path: "/invite/my",
  //   component: "@/pages/home/invite/my",
  //   title: "我的邀请",
  // },
  { path: "/findBot", component: "@/pages/home/findBot", title: "发现机器人" },
  { path: "/more", component: "@/pages/home/more", title: "更多活动" },
  { path: "/article", component: "@/pages/home/article/index" },
  { path: "/article/earn", component: "@/pages/home/article/earn" },
  { path: "/article/apply", component: "@/pages/home/article/apply" },
  { path: "/article/my", component: "@/pages/home/article/my" },
  { path: "/manager/article", component: "@/pages/manager/article" },

  {
    path: "/trade",
    component: "@/pages/home/tradeList",
    title: "持仓币种交易",
  },
  { path: "/trade/:id", component: "@/pages/home/trade" },
  { path: "/transfer/:id", redirect: "/trade/:id" }, //兼容处理

  { path: "/manager", component: "@/pages/manager/index" }, // 管理员首页
  { path: "/asset/deposit", component: "@/pages/manager/asset/assetChange" }, // 充值页面
  { path: "/asset/withdrawal", component: "@/pages/manager/asset/assetChange" }, // 提现页面
  { path: "/snapshots/:id", component: "@/pages/manager/asset/snapshot" }, // 资产记录页面

  { path: "/broadcast", component: "@/pages/manager/broadcast" },
  { path: "/red/pre", component: "@/pages/red/pre" }, // 群发红包
  { path: "/red", component: "@/pages/red/red" }, // 群发红包
  { path: "/red/timing", component: "@/pages/red/timing" }, // 群发红包
  { path: "/red/timingList", component: "@/pages/red/timingList" }, // 群发红包

  // { path: "/redRecord", component: "@/pages/redRecord" }, // 红包记录

  { path: "/create", component: "@/pages/create/index" }, // 创建社群
  { path: "/create/coin", component: "@/pages/create/coin" }, // 设置持仓币种
  { path: "/create/check", component: "@/pages/create/check" }, // 设置检查间隔

  // { path: "/statistics", component: "@/pages/manager/statistics" }, // 数据统计
  // { path: "/group", component: "@/pages/manager/group" }, // 群组管理
  // { path: "/group/setting", component: "@/pages/manager/groupSettings" }, // 群组设置

  // { path: "/board", component: "@/pages/manager/board" }, //  公告管理
  { path: "/broadcast/send", component: "@/pages/manager/sendBroadcast" }, // 群发公告

  // { path: "/setting", component: "@/pages/setting/index" }, //
  { path: "/setting/group", component: "@/pages/setting/group" },
  { path: "/setting/hello", component: "@/pages/setting/hello" },
  { path: "/setting/manager", component: "@/pages/setting/manager" },
  { path: "/setting/member", component: "@/pages/setting/member" },
  { path: "/setting/black", component: "@/pages/setting/member" },
  // { path: "/setting/invite", component: "@/pages/setting/invite" },
]
