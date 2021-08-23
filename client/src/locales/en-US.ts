import { getI18n } from "@/locales/tools"

const _i18n = {
  site: {
    title: "SuperGroup"
  },
  pre: {
    create: {
      title: "Create community",
      desc:
        "Community assistant could automatically create new groups, regularly check wallet balances, send red envelopes, and make announcements and other functions.",
      button: "Pay 0.01 XIN to active",
      action: "Create community",
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
    acceptTips: "Stop receiving group messages, not affect group announcements, all-important updates or news will be sent by announcement!",
    newNotice: "Reminder for new participant joined",
    receivedFirst: "Please recieve group messages first",
    auth: "Re-authorize",
    authConfirm: "Confirm to re-authorize?",
    exit: "Quit the group",
    exitConfirm: "Confirm to quit the group?",
    cancel: {
      title: "Stop receiving group messages",
      content: "Stop receiving group messages, not affect group announcements, all-important news will be sent by announcement!<br /> Enter the numbers below in order to confirm the operation."
    },
    exited: "Group quited",
    exitedDesc: "You have successfully exited the group, all data related to your account has been deleted, click the top right corner to close the page, hope you are back soon!"
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
    title: "Announcement management",
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
    totalMsg: "Total messages"
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
      confirmMute: "Confirm to mute {full_name}({identity_number}) for {mute_time} hours?",
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
      people: "user"
    },
    done: "End",
    center: "Membership",
    level0: "Free position membership",
    level0Desc: "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,1-Send chat messages,1-Premium membership could send images/videos and various types of messages",
    level0Sub: "Periodically check your wallet position; you could open the position membership freely depends on the balance is over {lamount} {symbol} to be a primary position membership or over {hamount} {symbol} to get the premium position membership.",
    level1: "Nonmembership",
    level1Desc: "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,0-Send chat messages",
    level2: "Primary Membership",
    level2Auth: "Primary position membership",
    level2Pay: "Primary paid membership",
    level2Desc: "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,1-Send chat messages,1-Send 10 messages per minute,1-You could send 3 types of messages/ such as text",
    level2Sub: "You could send 3 types of messages such as text, and send 5 to 10 messages per minute.",
    level5: "Premium membership",
    level5Auth: "Premium position membership",
    level5Pay: "Premium paid membership",
    level5Desc: "1-Receive all chat messages,1-Participate in grabbing Lucky Coin,1-Send chat messages,1-Send 20 messages per minute,1-You could send 9 types of messages/ such as text",
    level5Sub: "You could send 9 types of messages, and send 10 to 20 messages per minute.",
    upgrade: "Upgrade Membership",
    levelPay: "Pay {amount} {symbol} to get 1-year {level}. You could send {category} types of messages, such as text, with a sending limitation of {min} to {max} messages per minute.",
    checkPaid: "Check payment",
    authTips: "Opening for free via authorization will periodically check your wallet if your assets meet the position requirements, please see the documentation for more instructions: <a href='https://mixinmessenger.zendesk.com/hc/en-us/articles/4404710355220' target='_blank' rel='noopener noreferrer nofollow'>https://mixinmessenger.zendesk.com/hc/en-us/articles/4404710355220</a>",
    forFree: "Authorize access for free membership",
    forPay: "Pay {amount} {symbol} to get membership",
    expire: "Your membership expires on {date}, please renew it after expiration.",
    failed: "Membership opening failed",
    failedDesc: "Your balance does not meet the requirements to open the membership. Would you mind making sure to allow the authorization to read your assets? If your asset is stored in ExinOne's liquid pool, please open the asset page on the ExinOne bot and click the settings icon at the top, then switch on the authorization."
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
      forbidDesc2: "You're banned from the group. To join the group, please contact the admin.",
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
      downloadXinsheng: "Install JustChat",
    },

    search: {
      name: "Group name",
      holder: "Requirement for sending",
      or: "or",
      people: "Members",
    },
  },
  airdrop: {
    success: "Claimed, the reward will be transferred to you later",
    failed: "Sorry, you don't meet the requirement to claim the reward",
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
    open: "Chatting",
    article: "Information",
    more: "More",
    noActive: "No Data",
    noNews: "No Data",


    joinSuccess: "Joined Successfully",
    enterChat: "Start Chatting",
    enterHome: "Go to community homepage"
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
      cancelTop: "unpin"
    },
    confirmTop: "Confirm to pin",
    confirmCancelTop: "Confirm to unpin",

    liveReplay: {
      title: "Replay",
      delete: "Delete"
    },
    stat: {
      title: "Live statistics",
      read_count: "Views",
      deliver_count: "delivered",
      duration: "Duration（minutes）",
      user_count: "Number of participants（video）",
      msg_count: "Number of messages（video）"
    }
  },

  invite: {
    title: "Invite to join the group",
    desc: "Group invitation",
    card: "Send invitation card",
    link: "Copy invitation link",
    tip1: "Admin could add more admins through member management page",
    tipNotOpen: "The invitation bonus is disabled in the current group",
    tipOpen:
      "The invitation bonus is enable, the bonus will be sent as a Red Envelope, please claim it with 48 hours, the expired red envelopes cannot be claimed.<br /><br />" +
      "Please send the invitation card or invitation link to your Mixin contact directly!!! The invitation bonus is working Only when the invitee opens the card or clicks the link in the private conversation. If you send it to groups, bots, or web browser, the invitation is invalid!<br /><br />" +
      "Please Do Not disturb strangers, if you get too many reports, you may lose the invitation bonus or be banned from invitation qualified.",

    my: {
      title: "My invitations",
      reward: "Invitation bonus",
      people: "Invitees",
      "0": "Wait for qualified",
      "1": "Valid invitation",
      "2": "Valid invitation",

      noTitle: "Disable",
      noTips: "The invitation bonus will be enabled soon, the previous invitation is still valid.",

      noInvited: "No invitation",
      rule: "Check invitation rules",
    },
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

  //
  // manager: {
  //   members: "Total number of users",
  //   broadcasts: "Total number of announcement",
  //   conversations: "Total number of groups",
  //   list: "Net users",
  //
  //   asset: {
  //     title: "Assets",
  //     total: "Total assets",
  //     deposit: "Deposit",
  //     withdrawal: "Withdrawal",
  //     packet_send: "Send Red Envelope",
  //     packet_refund: "Red envelope refound",
  //     airdrop: "Airdrop",
  //     exin_otc: "community commission",
  //
  //     action: {
  //       deposit: "Pay",
  //       withdrawal: "Withdrawal",
  //     },
  //     checking: "Checking payment status",
  //     depositSuccess: "Payment successful",
  //     withdrawalSuccess: "Withdrawal successful",
  //   },
  // },

  modal: {
    check: "Checking payment result",
    loading: "Loading",
  },

  action: {
    tips: "Tips",
    cancel: "Cancel",
    save: "Save",
    confirm: "Confirm",
    submit: "Submit",
    continue: "Continue",
    know: "Got it",
  },

  success: {
    copy: "Copied",
    send: "Sent",
    operator: "Confirmed",
    save: "Saved",
    modify: "Modified"
  },
  error: {
    people: "Wrong for the user number",
    amount: "Wrong for the ammount",
    mixin: "Please open it with Mixin Messenger",
    empty: "Cannot be empty",
    modify: "Modification failed"
  },
}

const i18n = {}
getI18n(_i18n, i18n)
export default i18n
