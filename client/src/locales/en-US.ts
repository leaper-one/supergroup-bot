import { getI18n } from "@/locales/tools"

const _i18n = {
  site: {
    title: "SuperGroup"
  },
  pre: {
    create: {
      title: "Active Community assitant",
      desc:
        "Community assistant could automatically create new groups, regularly check wallet balances, send red envelopes, and make announcements and other functions.",
      button: "Pay 0.01 XIN to active",
      action: "Active Community assistant",
    },
    explore: {
      title: "Community assistant",
      desc:
        "Please add Community assistant as a group participant, and set it as an admin, then the community management is active.",
      button: "Find communities",
    },
  },

  setting: {
    title: "Setting"
  },
  join: {
    title: "Find communities",
    received: "Claim successfully",

    main: {
      join: "Authorize to access the group",
      joinTips:
        "Please authorize as it requires to check balance and other info.",

      appointBtn: "Subscribe",
      appointedBtn: "Subscribed",
      appointedTips: "Add to contact",

      receiveBtn: "Claim Airdrop",

      receivedBtn: "Claimed Airdrop",
      receivedTips: "Join Airdrop group",

      noAccess: "Not qualified",

      appointOver: "Airdrop is over",
    },

    modal: {
      auth: "Authorization failed",
      authDesc:
        "Please agree to authorize access to your assets, the data will only be used for balance checking.",
      authBtn: "Re-authorize",

      forbid: "Banned from group",
      forbidDesc1:
        "If you cannot join the group within 24 hours, please contact the admin or wait 24 hours before retrying to enter the group.",
      forbidDesc2:
        "You're banned from the group. To join the group, please contact the admin.",
      forbidBtn: "Got it.",

      shares: "Balance check failed",
      sharesBtn: "Recheck",
      sharesFail: "Failed",
      sharesTips: "Buy now",
      sharesCheck:
        "Please make sure your balance check has to meet at least one requirement below:",
      sharesCheck1: "No less than",
      sharesCheck2: "Balance check includes Mixin wallet.",

      appoint: "Subscribe successfully",
      appointDesc:
        "Thanks for subscribing! Please add this bot as a contact and turn on notification permission to receive the Airdrop qualification push alert.",
      appointBtn: "Add to contact",
      appointTips:
        "Tap the top right corner to close the bot, then wait for notification",
      receive: "Airdrops",
      receiveDesc:
        "Congratulations on getting MobileCoin Airdrop! Thank you for your support of MobileCoin, and welcome to claim the Airdrop and join the group.",
      receiveBtn: "Claim MOB Airdrop",
      receivedDesc:
        "This is the Airdrop for participating {comment}！ Thanks for your support of MobileCoin and Mixin!",
      receivedBtn: "Claimed MOB Airdrop",
    },

    code: {
      invite: "Scan with Mixin to {action}",
      download: "Download Mixin Messenger",
      action: {
        appoint: "subscribe",
        join: "join the group",
      },
    },

    search: {
      name: "Group name",
      holder: "Balance requirement",
      or: "or",
      people: "Members",
    },
  },

  home: {
    title: "Super Community",

    trade: "Trade",
    invite: "Invite",
    findGroup: "Find communities",
    findBot: "Find bots",
    more: "More events",
  },

  red: {
    title: "Group Red Envelopes",
    type: {
      title: "Types of Red Envelopes",
      "0": "Random Amount",
      "0Desc": "Depends on how lucky you are, grab more than others",
      "1": "Identical Amount",
      "1Desc": "Equal amount to everyone who opens it",
    },

    people: "Quantity",
    memo: "Best wishes",
    timingTitle: "Timed Red Envelopes",
    packetTime: "Time of sending",
    times: "Number of times",

    send: "Send Red Envelope",
    next: "Next step",
    addTiming: "Add Timed Red Envelope",
    tips:
      "The Red Envelope will be allocated proportionally to the number of participants in the group.",

    amount: "Amount",
    amountDesc: "Amount each",

    rate: "Red Envelope ratio",
    rateDesc:
      "There're {people} participants in the group，each time a red envelope is sent out, {rate}% of the group participants, about {receive} people can grab it, please note that the number of Red Envelopes is changing with the number of group participants.",

    timing: {
      title: "Time of Red Envelopes",
      time: "Time",
      morning: "a.m.",
      afternoon: "p.m.",
      hour: "Hour",
      minute: "minute",
      repeat: "Repeat",
      everyday: "Everyday",
      "0": "Sunday",
      "1": "Monday",
      "2": "Tuesday",
      "3": "Wednesday",
      "4": "Thursday",
      "5": "Friday",
      "6": "Saturday",
    },

    week: {
      "0": "日",
      "1": "一",
      "2": "二",
      "3": "三",
      "4": "四",
      "5": "五",
      "6": "六",
    },
  },

  invite: {
    title: "Invite to join group",
    desc: "Invitation",
    card: "Send invitation card",
    link: "Copy invitation link",
    tip1:
      "The group admin could add group participants via participant management",
    tipNotOpen: "The invitation bonus is disabled in the current group",
    tipOpen:
      "The invitation bonus is enable, the bonus will be sent as a Red Envelope, please claim it with 48 hours, the expired red envelopes cannot be claimed.<br /><br />" +
      "Please send the invitation card or invitation link to your Mixin contact directly!!! The invitation bonus is working Only when the invitee opens the card or click the link in the private conversation. If you send it to groups, bots, or web broswer, the invitation is invalid!<br /><br />" +
      "Please Do Not disturb strangers, if you get too many reports, you may lose the invatation bonus or banned from invitation qualified.",

    my: {
      title: "My invitations",
      reward: "Invitation bonus",
      people: "Invitees",
      "0": "Wait for qualified",
      "1": "Valid invitation",
      "2": "Valid invitations",

      noTitle: "Disable",
      noTips:
        "The invitation bonus will be enable soon, the previous invitation is still valid.",

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

    order: "Max {amount} {symbol}",

    maker: "Swap",
    taker: "{exchange}",

    Huobi: "Huobi",
    BigONE: "BigONE",
    Binance: "Binance",
    ExinSwap: "ExinSwap",
    exchange: "Exchange",
    MixSwap: "MixSwap",
    sign: "Multisig",
  },

  modal: {
    check: "Checking the payment result",
    loading: "Loading",
  },

  action: {
    tips: "Hint",
    cancel: "Cancel",
  },

  success: {
    copy: "Copied",
  },
  error: {
    people: "The quantity is wrong",
    amount: "The amount is wrong",
    mixin: "open it in Mixin Client",
  },
}
const i18n = {}
getI18n(_i18n, i18n)
export default i18n
