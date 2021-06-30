export default [
  { path: "/", component: "@/pages/home/index" },
  { path: "/auth", component: "@/pages/auth" },
  // { path: "/about", component: "@/pages/about" },
  { path: "/pre", component: "@/pages/pre/index" }, // 从机器人打开首页
  { path: "/join/:number", component: "@/pages/pre/join" }, // 申请加入持仓群页面
  { path: "/join", component: "@/pages/pre/join" }, // 申请加入持仓群页面
  { path: "/explore", component: "@/pages/pre/search", title: "发现社群" }, // 探索其他社群页面
  { path: "/pre/search", redirect: "/explore" }, // 兼容处理

  { path: "/home", component: "@/pages/home/index" },
  { path: "/activity", component: "@/pages/home/activity" },
  { path: "/setting", component: "@/pages/home/setting" },
  { path: "/exit", component: "@/pages/home/exit" },
  { path: "/news", component: "@/pages/home/news/index" },
  { path: "/news/addLive", component: "@/pages/home/news/addLive" },
  { path: "/news/liveDesc", component: "@/pages/home/news/liveDesc" },
  { path: "/news/liveReplay", component: "@/pages/home/news/liveReplay" },
  { path: "/news/liveStat", component: "@/pages/home/news/liveStat" },

  { path: "/manager/setting", component: "@/pages/setting/manager" },
  { path: "/manager/setting/base", component: "@/pages/setting/Base" },
  { path: "/manager/hello", component: "@/pages/setting/hello" },

  { path: "/broadcast", component: "@/pages/manager/broadcast" },
  { path: "/broadcast/send", component: "@/pages/manager/sendBroadcast" }, // 群发公告
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
  // { path: "/article", component: "@/pages/home/article/index" },
  // { path: "/article/earn", component: "@/pages/home/article/earn" },
  // { path: "/article/apply", component: "@/pages/home/article/apply" },
  // { path: "/article/my", component: "@/pages/home/article/my" },
  // { path: "/manager/article", component: "@/pages/manager/article" },

  { path: "/trade/:id", component: "@/pages/home/trade" },
  { path: "/transfer/:id", redirect: "/trade/:id" }, //兼容处理

  // { path: "/manager", component: "@/pages/manager/index" }, // 管理员首页
  // { path: "/asset/deposit", component: "@/pages/manager/asset/assetChange" }, // 充值页面
  // { path: "/asset/withdrawal", component: "@/pages/manager/asset/assetChange" }, // 提现页面
  // { path: "/snapshots/:id", component: "@/pages/manager/asset/snapshot" }, // 资产记录页面

  // { path: "/red/pre", component: "@/pages/red/pre" }, // 群发红包
  // { path: "/red", component: "@/pages/red/red" }, // 群发红包
  // { path: "/red/timing", component: "@/pages/red/timing" }, // 群发红包
  // { path: "/red/timingList", component: "@/pages/red/timingList" }, // 群发红包
  // { path: "/redRecord", component: "@/pages/redRecord" }, // 红包记录

  // { path: "/create", component: "@/pages/create/index" }, // 创建社群
  // { path: "/create/coin", component: "@/pages/create/coin" }, // 设置持仓币种
  // { path: "/create/check", component: "@/pages/create/check" }, // 设置检查间隔

  // { path: "/setting", component: "@/pages/setting/index" }, //
  // { path: "/setting/group", component: "@/pages/setting/group" },
  // { path: "/setting/hello", component: "@/pages/setting/hello" },
  // { path: "/setting/manager", component: "@/pages/setting/manager" },
  // { path: "/setting/member", component: "@/pages/setting/member" },
  // { path: "/setting/black", component: "@/pages/setting/member" },
  // { path: "/setting/invite", component: "@/pages/setting/invite" },
]
