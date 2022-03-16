import { getI18n } from "@/locales/tools"

const _i18n = {
  site: {
    title: "SuperGroup"
  },
  pre: {
    create: {
      title: "Open community",
      desc: "Community assistant could automatically create new groups, regularly check wallet balances, send red envelopes, and make announcements and other functions.",
      button: "Pay 0.01 XIN to active",
      action: "Create communit",
    },
    explore: {
      title: "Community assistant",
      desc: "Please add Community assistant as a group participant, and set it as an admin, then the community management is active.",
      button: "Find communities",
    },
  },

  setting: {
    title: "Settings",
    accept: "Receive group messages",
    acceptTips:
      "Stop receiving group messages, not affect group announcements, all-important updates or news will be sent by announcement!",
    newNotice: "Reminder for new participant joined",
    useProxy: "Use proxy",
    receivedFirst: "Please recieve group messages first",
    auth: "Re-authorize",
    authConfirm: "Confirm to re-authorize?",
    exit: "Quit the group",
    exitConfirm: "Confirm to quit the group?",
    cancel: {
      title: "Stop receiving group messages",
      content:
        "Stop receiving group messages, not affect group announcements, all-important news will be sent by announcement!<br /> Enter the numbers below in order to confirm the operation.",
    },
    exited: "Group quited",
    exitedDesc:
      "You have successfully exited the group, all data related to your account has been deleted, click the top right corner to close the page, hope you are back soon!",
  },

  manager: {
    setting: "Settings",
    base: "General settings",
    description: "Group profile",
    welcome: "Welcome message",
    high: "Advanced management",

    helloTips: "Welcome message shall be sent to newly joined members, while the other group members cannot see it.",
  },
  broadcast: {
    a: "Announcement",
    title: "Annoucement management",
    holder: "Please post the announcement",
    recall: "Recall",
    confirmRecall: "Confirm to recall",
    recallSuccess: "Recall successfully",
    status0: "Sending",
    status1: "Sent",
    status2: "Recalling",
    status3: "Recalled",
    checkNumber: "Please check the numbers you entered are the same",
    sent: "Group announcement",
    input: "Enter the numbers above in order to send the announcement in the group",
    fill: "Please write the announcement first",
    send: "Send",
  },
  stat: {
    title: "Statistics",
    totalUser: "Total users",
    highUser: "Minimum position users",
    weekUser: "Weekly new users",
    weekActiveUser: "Weekly Active Users",
    totalMessage: "Total number of messages",
    weekMessage: "Weekly messages",
    all: "All",
    month: "Recent month",
    week: "Recent week",
    user: "Users",
    newUser: "New users",
    activeUser: "Active users",
    msg: "Messages",
    dailyMsg: "Daily messages",
    totalMsg: "Total messages",
  },
  member: {
    title: "User management",
    status8: "Lecture",
    status9: "Admin",
    hour: "Active {n} hours ago",
    day: "Active {n} days ago",
    month: "Active {n}  months ago",
    year: "Active {n} years ago",
    action: {
      set: "Set to {c}",
      cancel: "Remove from {c}",
      confirmSet: "Confirm to set {full_name}({identity_number}) to {c} ?",
      confirmCancel: "Confirm to remove {full_name}({identity_number}) from {c} ?",
      guest: "Lecture",
      admin: "Admin",
      mute: "Mute",
      confirmMute:
        "Confirm to mute {full_name}({identity_number}) for {mute_time} hours?",
      block: "Block",
      confirmBlock: "Confirm to block {full_name}({identity_number})?",
    },
    modal: {
      unit: "Hours",
      desc: "The user will be muted for 1 hour, muted status won't affect to receive messages and grab red envelopes."
    },
    status: {
      title: "User's type",
      all: "All",
      guest: "Lecture",
      admin: "Admin",
      mute: "Muted",
      block: "Blocked",
      people: "user",
    },
    done: "End",

    center: "Membership",
    level0: "Free position membership",
    level0Desc:
      "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,1-Send chat messages,1-Premium membership could send images/videos and various types of messages",
    level0Sub:
      "Periodically check your wallet position; you could open the position membership freely depends on the balance is over {lamount} {symbol} to be a primary position membership or over {hamount} {symbol} to get the premium position membership.",
    level1: "Nonmembership",
    level1Desc: "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,0-Send chat messages",
    level2: "Primary Membership",
    level2Auth: "Primary position membership",
    level2Pay: "Primary paid membership",
    level2Desc:
      "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,1-Send chat messages,1-Send 10 messages per minute,1-You could send 3 types of messages/ such as text.",
    level2Sub: "You could send 3 types of messages such as text, and send 5 to 10 messages per minute.",
    level5: "Premium membership",
    level5Auth: "Premium position membership",
    level5Pay: "Premium paid membership",
    level5Desc:
      "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,1-Send chat messages,1-Send 20 messages per minute,1-You could send 9 types of messages/ such as text",
    level5Sub: "You could send 9 types of messages, and send 10 to 20 messages per minute.",

    upgrade: "Upgrade Membership",
    levelPay:
      "Pay {amount} {symbol} to get 1-year {level}. You could send {category} types of messages, such as text, with a sending limitation of {min} to {max} messages per minute.",
    checkPaid: "Check payment",
    authTips:
      "Opening for free via authorization will periodically check your wallet if your assets meet the position requirements, please see the documentation for more instructions:<a href='https://w3c.group/c/1628159023237756'>https://w3c.group/c/1628159023237756</a>",
    forFree: "Authorize access for free membership",
    forPay: "Pay {amount} {symbol} to get membership",

    cancel: "Revoke membership",
    cancelDesc:
      "You will lose the membership when you click on the revoke button below and the group bot will not be able to read your asset information anymore. You could re-authorize the bot anytime to get the membership back.",

    expire: "Your membership expires on {date}, please renew it after expiration.",
    failed: "Membership opening failed",
    failedDesc:
      "Your balance does not meet the requirements to open the membership. Would you mind making sure to allow the authorization to read your assets? If your asset is stored in ExinOne's liquid pool, please open the asset page on the ExinOne bot and click the settings icon at the top, then switch on the authorization.",
  },
  advance: {
    title: "Advanced Settings",
    mute: "All Mute",
    muteConfirm: "Are all {action} mute all {tips}?",
    muteTips: "（except administrators and guests）",
    open: "Active",
    close: "Inactive",
    newMember: "New participant join Reminder",
    newMemberConfirm: "Are you sure {action} new participant join reminder?",
    sliderConfirm: "Slide to confirm operation",
    proxy: "Prohibit mutual contact",
    proxyConfirm: "Are you sure {action} prohibit mutual contact?",
    proxyTips: "（except administrators）",
    msgAuth: "Message Permissions",
    member: {
      1: "Nonmembership",
      2: "Primary membership",
      5: "Premium membership",
      tips: "{status} can send up to {count} messages per minute."
    },
    plain_text: "Text",
    plain_sticker: "Stickers",
    plain_image: "Picture",
    plain_video: "Video",
    lucky_coin: "Red envelop",
    plain_post: "Post",
    plain_live: "Live card",
    plain_contact: "Bot Contact",
    plain_transcript: "Chat history",
    plain_data: "File",
    url: "Link",
    app_card: "App Card",
  },
  join: {
    title: "Find communities",
    received: "Claim successfully",
    open: "Open in Mixin",

    main: {
      join: "Authorize to join",
      joinTips: "【Risk warning】 Mixin does not endorse or guarantee any token price nor the project.",

      appointBtn: "Subscribe",
      appointedBtn: "Subscribed",
      appointedTips: "Add contact",

      member: "Members",

      receiveBtn: "Receive Airdrop",

      receivedBtn: "Claimed Airdrop",
      receivedTips: "Join Airdrop group",

      noAccess: "Not qualified",

      appointOver: "Airdrop is over",
    },

    modal: {
      auth: "Authorization failed",
      authDesc: "Please agree to authorize access to your assets, the data will only be used for balance checking.",
      authBtn: "Re-authorize",

      forbid: "Banned from group",
      forbidDesc1: "If you cannot join the group within 24 hours, please get in touch with the admin or wait 24 hours before retrying to enter the group.",
      forbidDesc2: "You're banned from the group. To join the group, don't hesitate to get in touch with the admin.",
      forbidBtn: "Got it.",

      shares: "Balance check failed",
      sharesBtn: "Recheck",
      sharesFail: "Failed",
      sharesTips: "Buy now",
      sharesCheck: "Please make sure your balance check has to meet at least one requirement below:",
      sharesCheck1: "No less than",
      sharesCheck2:
        "The balance check include Mixin wallet, Exin 流动池， 活期宝， 省心投 and Fox's Defi products such as flex-term, fixed-term, regular Invest, and Node.",

      appoint: "Subscribe successfully",
      appointDesc:
        "Thanks for subscribing! Please add this bot as a contact and turn on notification permission to receive the Airdrop qualification push alert.",
      appointBtn: "Add to contact",
      appointTips: "Tap the top right corner to close the bot, then wait for notification",
      receive: "Airdrops",
      receiveDesc:
        "Congratulations on getting MobileCoin Airdrop! Thank you so much for being so supportive of MobileCoin, and welcome to claim the Airdrop and join the group.",
      receiveBtn: "Claim MOB Airdrop",
      receivedDesc:
        "This is the Airdrop for participating {comment}！ Thanks for your support of MobileCoin and Mixin!",
      receivedBtn: "Claimed MOB Airdrop",
    },

    code: {
      invite: "Scan to join the group with Mixin Messenger",
      download: "Download Mixin Messenger",
      href: "https://mixin.one/messenger",
    },

    search: {
      name: "Group name",
      holder: "Requirement for sending",
      or: "or",
      people: "Members",
    },
  },
  mint: {
    join: "Participate in yield farming",
    receive: "Claim rewards",
    first: "头矿",
    time: "Period",
    reward: "Rewards",
    theme: "Event & Rewards",
    part: "Divide everyday",
    and: "and",
    duration: "{aY} 年 {aM} 月 {aD} 日至 {bY} 年 {bM} 月 {bD} 日（{d}天）",
    receiveTime: "Claim date",
    receiveTimeTips: "Claim reward in the next day, please go bakc to claim it by yourself after 10:00 am UTC+8 next day.",
    daily: "Daily farming",
    faq: "FAQ",
    continue: "Continue",
    close: "Close window",
    auth: "Authorize to read the assets, otherwise you cannot participate in the yield farming event.",
    pending: "You haven't participated in the yield farming yet, please back to claim the rewards after you add the liquidity.",

    record: {
      title: "Claim records",
      0: "All",
      1: "unclaimed",
      2: "Claimed",
      3: "Absent",
      pair: "Trading pair",
      lp: "LP amount",
      per: "Ratio",
      tips: "The rewards will be distributed within 2 hours, please check them on the wallet.",
      wait: "Please do not duplicate rewards"
    }
  },
  airdrop: {
    success: "Claimed, the reward will be transferred to you later",
    failed: "Sorry, you don't meet the requirement to claim the reward",
    assetCheck: "Please note that you need to have more than ${amount} on the wallet to claim the rewards",
  },

  home: {
    title: "Community assistant",
    people_count: "Members",
    week: "This week",
    trade: "Trade",
    invite: "Invite",
    findGroup: "Find communities",
    findBot: "Find bots",
    activity: "Events",
    redPacket: "LuckyCoin",
    reward: "Give Tips",
    claim: "Sign in",
    open: "Chatting",
    article: "Information",
    more: "More",
    noActive: "No Data",
    noNews: "No Data",
    notStart: "The event doesn't start",
    isEnd: "The event is over",

    joinSuccess: "Joined Successfully",
    enterChat: "Start Chatting",
    enterHome: "Go to community homepage",

    findBotURL: "https://bots.mixin.zone",
  },

  news: {
    all: "All",
    replay: "Replay",
    broadcast: "Annoucement",
    sendBroadcast: "Send announcement",
    sendLive: "Add live streaming preview",
    live: "Live streaming",
    confirmStart: "Confirm to start live streaming",
    confirmEnd: "Confirm to end live streaming",
    prompt: "Please enter the link of the live streaming",
    form: {
      img: "Live banner",
      category: "Type of live streaming",

      "1": "Video",
      "2": "Text-image",

      user: "Live streaming guest",
      title: "Tile",
      desc: "Description",
    },
    livePreview: "Live streaming preview",
    action: {
      stop: "Stop live streaming",
      delete: "Delete",
      edit: "Edit preview",
      share: "share preview",
      start: "Start live",
      top: "Pin",
      cancelTop: "unpin",
    },
    confirmTop: "Confirm to pin",
    confirmCancelTop: "Confirm to unpin",

    liveReplay: {
      title: "Replay",
      delete: "Delete",
    },
    stat: {
      title: "Live statistics",
      read_count: "Views",
      deliver_count: "delivered",
      duration: "Duration（minutes）",
      user_count: "Number of participants（video）",
      msg_count: "Number of messages（video）",
    },
  },

  invite: {
    title: "Invite to join the group",
    desc: "Group invitation",
    card: "Send invitation card",
    link: "Copy invitation link",
    tip1: "Admin could add more admins through member management page",
    tip2: "The invitation bonus has been opened! Invite your friends to join any community and participate in the sign-in and lucky draw to get energy rewards.",
    tip3: "Please do not harass strangers! Once it is detected that the current user has been reported too many times, no qualification to receive rewards.",
    my: {
      title: "My invitations",
      reward: "Invitation bonus",
      people: "Invitees",
      noInvited: "No invitation",
      rule: "Check invitation rules",
    },
    claim: {
      title: "Get energy reword if you invite any friend join the community.",
      btn: "Invite",
      count: "Invitees {count} "
    }
  },

  transfer: {
    title: "Trade {name}",
    price: "Price",
    pool: "Pool",
    earn: "24H AROR",
    amount: "24H Volume",
    method: "Trading method",
    noPrice: "No price",

    order: "Max order {amount} {symbol}",

    maker: "AMM",
    taker: "Agent buy from{exchange}",

    Huobi: "Huobi",
    BigONE: "BigONE",
    Binance: "Binance",
    ExinSwap: "ExinSwap",
    MixSwap: "MixSwap",
    exchange: "Exchange",
    sign: "Individuals Multisig trading",

    coin: "Trade with crypto",
    otc: "OTC",

    auth: "Blue sheild Trust",
    identity: "KYC",

    payMethod: "Payment method",
    bank: "Back",
    alipay: "Alipay",
    wechatpay: "WeChat",

    category: "Token",
    limit: "Trade limite",
    in5minRate: "5 minute completion rate",
    orderSuccessRank: "Order completion rate",
    multisigOrderCount: "Total number of orders",
  },

  reward: {
    title: "Give Tips",
    who: "To whom",
    amount: "Amount",
    less: "At least $1",
    success: "Done",
    isLiving: "The gifting feature is disabled during the live streaming, and please retry after it ends.",
  },

  claim: {
    title: "Lucky Draw",
    tag: "Trial operation",
    receive: "Claim reward",
    receiveSuccess: "Claimed, you will get the reword later.",
    drew: "won",
    worth: "{prefix} $ {value}",
    now: " ",
    you: " ",
    ticketCount: "chances left",
    success: "Sign in successfully",
    ok: "Got it",
    join: "Join community",
    open: "Open community",

    energy: {
      title: "Energy",
      describe: "1 draw for every 100 energy",
      exchange: "Redeem",
      success: "Redeemed",
      checkin: {
        label: "Sign in",
        checked: "Signed in",
        count: "The week {count}/7",
        describe: "10 energy for daily sign-in, 50 extra energy for 5 days in 1 week",
      },
    },
    records: {
      title: "Records",
      winning: "Reward records",
      energy: "Energy records",
      lottery: "Lucky Draw",
      power_lottery: "Redeem to draw",
      power_claim: "Sign in everday",
      power_claim_extra: "Sign in 5 times this week",
      power_invitation: "Invitation bonus",
    },
  },
  guess: {
    name: "Guess the Price Win the {coin} Pirze",
    up: "higher",
    down: "lower",
    flat: "equal",
    sure: "Confirm",
    notsure: "Cancel",
    okay: "Okay",
    goChoose: "Option",
    todayGuess: "{coin} Guess today's price",
    todyDesc:
      "Today {time} UTC+8 {coin} token price is ${usd} (price collected from Coingecko.com) please guess tomorrow's {time} UTC+8:",

    choose: {
      tip: "Please note",
      info: "Please select an option before click OK button。",
    },
    confirm: {
      tip: "Please note",
      info: "you cannot change the answer after confirmation, are you sure the selection?",
    },
    success: {
      tip: "Congratulations",
      info: "You have participated the price guess event, the {coin} price result will be announced after {start} UTC+8.",
    },
    notstart: {
      tip: "It doesn't start",
      info: "Today's price guess doesn't start, please participate during this period {start} - {end} UTC+8.",
    },
    missing: {
      tip: "You missed it",
      info: "Today's price guess is over, please go back to participate tomorrow between {start} - {end} UTC+8.",
    },
    end: {
      tip: "The event is over",
      info: 'The price guess event is over, you could check the results on "my records" page.',
    },
    records: {
      name: "My records",
      history: "{coin} price guess records",
      consecutiveplay: "You participated consecutively",
      condition: "Participate for consecutive 3 days, you can participate in the reward divide",
      play: "You have enrolled in",
      guess: "Guess",
      up: "Up",
      down: "Down",
      flat: "No change",
      day: "day, ",
      vip: "become a valid participant, ",
      playresult: "Result:",
      date: "Date",
      result: "Results",
      win: "Win",
      lose: "lose",
      pending: "Pending",
      notplay: "Absent",
      notstart: "Not start",
    },
  },
  trading: {
    rule: "Trading rules",
    time: "Date",
    reward: "Trading rewards",
    auth: "Authorize to participate",
    viewRank: "Check ranking",
    modalDesc: "You can exchange USDT, BTC, ETH and other coins for {symbol} via MixSwap or 4swap to participate in the trading competition.",
    mixSwap: "Trade via MixSwap",
    swap: "Trade via 4swap",
    rank: "Ranking",
    ranked1: "The {i} place",
    amount: "Your trading volume is {amount} {symbol}",
    ranked2: "Ranked {i}",
    noRank: "Not in the top 10.",
  },
  modal: {
    check: "Checking the payment result",
    loading: "Loading",
  },

  action: {
    tips: "Hint",
    cancel: "Cancel",
    save: "Save",
    confirm: "Confirm",
    submit: "Submit",
    continue: "Continue",
    know: "Got it",
    open: "Open",
  },

  success: {
    copy: "Copied",
    send: "Sent",
    operator: "Done",
    save: "Saved",
    modify: "Edited",
  },
  error: {
    people: "The number error",
    amount: "The amount error",
    mixin: "Please open in Mixin",
    empty: "Cannot be empty",
    modify: "Edit failed",
  },
}
const i18n = {}
getI18n(_i18n, i18n)
export default i18n
