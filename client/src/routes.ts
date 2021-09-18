const routes = [
  { path: "/", component: "@/pages/home/index" },
  { path: "/auth", component: "@/pages/auth" },
  { path: "/join/:number", component: "@/pages/pre/join" }, // 申请加入持仓群页面
  { path: "/join", component: "@/pages/pre/join" }, // 申请加入持仓群页面
  { path: "/explore", component: "@/pages/pre/search", title: "join.title" }, // 探索其他社群页面
  { path: "/pre/search", redirect: "/explore" }, // 兼容处理

  { path: "/home", component: "@/pages/home/index" },
  { path: "/reward", component: "@/pages/home/reward" },
  { path: "/lottery", component: "@/pages/home/lottery" },
  { path: "/lottery/records", component: "@/pages/home/lottery/records" },
  { path: "/activity", component: "@/pages/home/activity" },
  { path: "/activity/:id", component: "@/pages/home/activity/guess" },
  { path: "/activity/:id/records", component: "@/pages/home/activity/records" },
  { path: "/setting", component: "@/pages/home/setting" },
  { path: "/exit", component: "@/pages/home/exit" },
  { path: "/news", component: "@/pages/home/news/index", title: "site.title" },
  { path: "/news/addLive", component: "@/pages/home/news/addLive" },
  { path: "/news/liveDesc", component: "@/pages/home/news/liveDesc" },
  { path: "/news/liveReplay/:id", component: "@/pages/home/news/liveReplay" },
  { path: "/news/liveStat", component: "@/pages/home/news/liveStat" },
  { path: "/member", component: "@/pages/home/member" },

  { path: "/manager/setting", component: "@/pages/manager/index" },
  { path: "/manager/setting/base", component: "@/pages/manager/base" },
  { path: "/manager/hello", component: "@/pages/manager/hello" },
  { path: "/manager/stat", component: "@/pages/manager/stat" },
  { path: "/manager/member", component: "@/pages/manager/member" },

  { path: "/broadcast", component: "@/pages/manager/broadcast" },
  { path: "/broadcast/send", component: "@/pages/manager/sendBroadcast" }, // 群发公告
  {
    path: "/invite",
    component: "@/pages/home/invite/index",
    title: "邀请入群",
  },
  {
    path: "/findBot",
    component: "@/pages/home/findBot",
    title: "home.findBot",
  },
  { path: "/more", component: "@/pages/home/more", title: "home.more" },

  { path: "/trade/:id", component: "@/pages/home/trade" },
  { path: "/transfer/:id", redirect: "/trade/:id" }, //兼容处理
]

export default routes.map((item) => ({ title: "site.title", ...item }))
