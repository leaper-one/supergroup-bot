import { getI18n } from "@/locales/tools"
import { $set } from "@/stores/localStorage"

$set("umi_locale", navigator.language.includes("zh") ? "zh-CN" : "en-US")

const _i18n = {
  site: {
    title: "Links公式アカウント",
  },
  pre: {
    create: {
      title: "オープンコミュニティ",
      desc: "コミュニティ・アシスタントは、新しいグループの自動作成、グループに投げ銭の送付、アナウンスなどを行うことができます。",
      button: "0.01XINでコミュニティを作成",
      action: "コミュニティの作成",
    },
    explore: {
      title: "コミュニティ・アシスタント",
      desc: "コミュニティ・アシスタントをグループに追加して管理者に設定し、コミュニティ管理機能を有効にしてください。",
      button: "コミュニティを探す",
    },
  },

  setting: {
    title: "設定",
    accept: "チャットメッセージの受信",
    acceptTips:
      "チャットの受信を停止しても、重要なプロジェクトのアナウンスはすべて掲載されます。",
    newNotice: "新規参加者への注意事項",
    useProxy: "プロキシIDの使用",
    receivedFirst: "まずはチャットでメッセージを受け取ってください",
    auth: "再認証",
    authConfirm: "再認証を行いましたか？",
    exit: "コミュニティから退会",
    exitConfirm: "コミュニティから退会しますか？",
    cancel: {
      title: "チャットの受付の停止",
      content:
        "チャットの受付を停止しても、重要なプロジェクトのアナウンスはすべて掲載されます。以下の数字を順番に入力して、操作を確定してください。",
    },
    exited: "コミュニティから退会しました",
    exitedDesc:
      "コミュニティから退会しました。右上をクリックしてコミュニティページを閉じてください。",
  },

  manager: {
    setting: "設定",
    base: "基本設定",
    description: "プロフィール",
    welcome: "ウェルカムメッセージ",
    high: "詳細設定",

    helloTips: "ウェルカムメッセージは、グループに新しく参加した人だけが見ることができ、他のユーザーは見ることができません。",
  },
  broadcast: {
    a: "アナウンス",
    title: "アナウンス設定",
    holder: "アナウンス内容を記入してください",
    recall: "キャンセル",
    confirmRecall: "本当に解除しますか？",
    recallSuccess: "解除しました",
    status0: "送信中",
    status1: "送信済み",
    status2: "解除処理中",
    status3: "解除済み",
    checkNumber: "数字が一致していることをご確認ください",
    sent: "グループへのアナウンス",
    input: "アナウンスを送信するには、上記の数字を入力してください",
    fill: "アナウンス内容を記入してください",
    send: "送信",
  },
  stat: {
    title: "統計情報",
    totalUser: "総ユーザー数",
    highUser: "指定トークン一定保有ユーザー数",
    weekUser: "新規ユーザー数/週",
    weekActiveUser: "アクティブユーザー数/週",
    totalMessage: "総メッセージ数",
    weekMessage: "メッセージ数/週",
    all: "全て",
    month: "直近1ヶ月",
    week: "直近1週間",
    user: "ユーザー",
    newUser: "新規ユーザー",
    activeUser: "アクティブユーザー",
    msg: "メッセージ",
    dailyMsg: "メッセージ数/日",
    totalMsg: "総メッセージ数",
  },
  member: {
    title: "メンバー管理",
    status8: "モデレーター",
    status9: "管理者",
    hour: "{n}時間前にアクティブ",
    day: "{n}日前にアクティブ",
    month: "{n}ヶ月前にアクティブ",
    year: "{n}年前にアクティブ",
    action: {
      set: "{c}に設定",
      cancel: "{c}から解除",
      confirmSet: "{full_name}({identity_number})を{c}に設定しますか？",
      confirmCancel: "{full_name}({identity_number})を{c}から解除しますか？",
      guest: "モデレーター",
      admin: "管理者",
      mute: "ミュート",
      confirmMute:
        "{full_name}({identity_number})を{mute_time}時間ミュートしますか？",
      block: "ブロック",
      confirmBlock: "{full_name}({identity_number})をブロックしますか？",
    },
    modal: {
      unit: "時間",
      desc: "ユーザーは指定時間ミュートされます。メッセージの受信や投げ銭の取得には影響しません。",
    },
    status: {
      title: "メンバー種類",
      all: "全て",
      guest: "モデレーター",
      admin: "管理者",
      mute: "ミュート",
      block: "ブロック",
      people: "メンバー",
    },
    done: "終了",

    center: "会員ランク",
    level0: "通貨保有会員",
    level0Desc:
      "1-全てのチャットの受信,1-投げ銭の受取,1-チャットへのメッセージ送信,1-1分間に最大20メッセージを送信可能,1-ゴールド会員は写真、ビデオ、他の種類のメッセージを送ることができます。",
    level0Sub:
      "定期的なウォレットチェックの際のウォレット内資産に応じて、会費無料で、通貨保有シルバー会員または通貨保有ゴールド会員に自動登録されます。ウォレット残高が{lamount} {symbol}以上の場合は通貨保有シルバー会員、{hamount} {symbol}以上の場合は通貨保有ゴールド会員になります。",
    level1: "レギュラー会員",
    level1Desc: "1-全てのチャットの受信,1-投げ銭の受取,0-チャットへのメッセージ送信",
    level2: "シルバー会員",
    level2Auth: "通貨保有シルバー会員",
    level2Pay: "有料シルバー会員",
    level2Desc:
      "1-全てのチャットの受信,1-投げ銭の受取,1-チャットへのメッセージ送信,1-1分間に最大10メッセージを送信可能,1-テキストなど3種類のメッセージを送信することができます。",
    level2Sub: "テキストを含む3種類のメッセージを、1分間最大10通送信できます。",
    level5: "ゴールド会員",
    level5Auth: "通貨保有ゴールド会員",
    level5Pay: "有料ゴールド会員",
    level5Desc:
      "1-全てのチャットを受信することができます,1-投げ銭の受取,1-チャットへのメッセージ送信,1-1分間に20メッセージを送信することができます,1-テキストを含む9種類のメッセージを送信することができます。",
    level5Sub: "テキストを含む9種類のメッセージを、1分間最大20通送信できます。",

    upgrade: "会員ランクのアップグレード",
    levelPay:
      "{amount} {symbol}を支払い、1年間{level}となります。テキストなど{category}種類のメッセージを送信でき、1分間に{min}～{max}通のメッセージの送信が可能です。",
    checkPaid: "支払確認",
    authTips:
      "ウォレット内資産による資産保有会員は、定期的に資産の確認が行われます。詳細な手順については、以下のドキュメントをご覧ください。：<a href='https://w3c.group/c/1628159023237756'>https://w3c.group/c/1628159023237756</a>",
    forFree: "無料で利用可能",
    forPay: "{amount} {symbol}を支払う",

    cancel: "会員の解除",
    cancelDesc:
      "会員の解除を行うと会員資格を失い、コミュニティシステムはあなたの資産情報を読み取ることができなくなります。",

    expire: "会員資格は{date}まで有効です。期限末までに更新してください。",
    failed: "会員ランクの有効化に失敗しました。",
    failedDesc:
      "あなたは会員ランク有効化の要件を満たしていません。資産の読み取りが許可されていることを確認してください。"
  },
  advance: {
    title: "詳細設定",
    mute: "グループ全体をミュート",
    muteConfirm: "グループ全体のミュートを{action}しますか？{tips}",
    muteTips: "（管理者とゲストは対象外です）",
    open: "開始",
    close: "終了",
    newMember: "グループ参加通知",
    newMemberConfirm: "グループ参加の通知を{action}しますか？",
    sliderConfirm: "スライドして操作を確認する",
    proxy: "ユーザー間の連絡先交換",
    proxyConfirm: "ユーザー間の連絡先交換を{action}しますか？",
    proxyTips: "（管理者は対象外です）",
    msgAuth: "メッセージの許可",
    member: {
      1: "レギュラー会員",
      2: "シルバー会員",
      5: "ゴールド会員",
      tips: "{status}が1分間に送信できるメッセージは最大{count}通です。"
    },
    plain_text: "文字",
    plain_sticker: "スタンプ",
    plain_image: "画像",
    plain_video: "動画",
    lucky_coin: "グループへの投げ銭",
    plain_post: "投稿",
    plain_live: "ライブ配信",
    plain_contact: "連絡先",
    plain_transcript: "チャット記録",
    plain_data: "ファイル",
    url: "リンク",
    app_card: "カード",
  },
  join: {
    title: "コミュニティを探す",
    received: "受取成功",
    open: "Linksで開く",

    main: {
      join: "参加",
      joinTips: "【注意】Linksがプロジェクトの発行するトークンの推奨、安全性の保証を行うものではありません。",

      appointBtn: "予約",
      appointedBtn: "予約済み",
      appointedTips: "連絡先の追加",

      member: "メンバー",

      receiveBtn: "エアドロップを受け取る",

      receivedBtn: "すでにエアドロップは受け取り済みです",
      receivedTips: "エアドロップに参加する",

      noAccess: "参加資格がありません",

      appointOver: "エアドロップは終了しました",
    },

    modal: {
      auth: "認証に失敗しました",
      authDesc: "資産へのアクセスを許可してください。こちらは会員資格の検証にのみ使用されます。",
      authBtn: "再認証",

      forbid: "グループへの参加を禁止する",
      forbidDesc1: "24時間以内にグループに参加できない場合は、管理者に連絡するか、24時間後に再参加してください。",
      forbidDesc2: "あなたはグループから退会させられました。グループに再参加するには、管理者に連絡してください。",
      forbidBtn: "わかりました",

      shares: "会員ランク情報の検出に失敗しました",
      sharesBtn: "再検出",
      sharesFail: "検出不可",
      sharesTips: "今すぐ購入する",
      sharesCheck: "以下の保有資産条件のいずれかを満たしていることをご確認ください。",
      sharesCheck1: "以上",
      sharesCheck2:
        "会員ランク情報の検出は、Linksウォレットに対応しています。",

      appoint: "予約が完了しました",
      appointDesc:
        "ご予約ありがとうございます。 エアドロップの通知を受け取るには、このミニアプリの通知をオンにしてください。",
      appointBtn: "連絡先の追加",
      appointTips: "右上をタップして通知をオフにする",
      receive: "エアドロップ報酬",
      receiveDesc:
        "MobileCoinエアドロップへの参加資格の取得、おめでとうございます。 MobileCoinのサポートに感謝し、エアドロップを受け取り、グループに参加することを歓迎します。",
      receiveBtn: "MOBを受け取る",
      receivedDesc:
        "{comment}に参加された方へのエアドロップ報酬です！いつもMobileCoinとLinksを応援して頂き、ありがとうございます。",
      receivedBtn: "受取済み",
    },

    code: {
      invite: "Linksでグループに参加する",
      download: "Linksをダウンロードする",
      downloadXinsheng: "JustChatをダウンロードする",
    },

    search: {
      name: "グループ名",
      holder: "ホルダー",
      or: "または",
      people: "メンバー",
    },
  },
  mint: {
    join: "マイニングに参加する",
    receive: "報酬を受け取る",
    first: "マイニングを始める",
    time: "イベント開催時間",
    reward: "イベント報酬",
    theme: "イベントと報酬",
    part: "毎日の採掘量",
    and: "と",
    duration: "{aY} 年 {aM} 月 {aD} 日から {bY} 年 {bM} 月 {bD} 日（{d}日間）",
    receiveTime: "受け取り時間",
    receiveTimeTips: "報酬は2日目に発行されますので、翌日22時（BST）以降にイベントページより手動で受け取りください。",
    daily: "デイリーマイニング",
    faq: "よくある質問",
    continue: "参加を継続する",
    close: "ポップアップを閉じる",
    auth: "資産の読み取りを許可してください。許可しない場合、報酬の受け取りに参加することができません。",
    pending: "まだマイニングキャンペーンに参加していない方は、報酬を受け取る前にマイニングキャンペーンに参加してください。",

    record: {
      title: "報酬の取得記録",
      0: "全て",
      1: "未入手",
      2: "入手済み",
      3: "不参加",
      pair: "取引ペア",
      lp: "LPトークンの数量",
      per: "売上高に対する比率",
      tips: "報酬は2時間以内に発行されます。詳細はウォレットをご確認ください。",
      wait: "二重請求はご遠慮ください。"
    }
  },
  airdrop: {
    success: "受取に成功しました。のちほど入金されます。",
    failed: "受取に失敗しました。受取条件を満たしていません。",
    assetCheck: "注意:エアドロップはウォレット残高が{amount}USD以上の場合にのみ請求可能です。",
  },

  home: {
    title: "コミュニティ・アシスタント",
    people_count: "コミュニティ<br/>人数",
    week: "今週",
    trade: "トレード",
    invite: "招待",
    findGroup: "コミュニティを探す",
    findBot: "ミニアプリを探す",
    activity: "イベント",
    redPacket: "グループに<br/>投げ銭",
    reward: "リワード",
    claim: "デイリー<br/>ボーナス",
    open: "チャット",
    article: "公式情報",
    more: "その他のイベント",
    noActive: "イベントはありません",
    noNews: "情報なし",
    notStart: "イベントは開始されていません",
    isEnd: "イベントは終了しました",

    joinSuccess: "参加しました",
    enterChat: "チャットに参加する",
    enterHome: "コミュニティホームへ",
  },

  news: {
    all: "すべて",
    replay: "再生",
    broadcast: "アナウンス",
    sendBroadcast: "アナウンスを送信",
    sendLive: "ライブの追加",
    live: "ライブ",
    confirmStart: "ライブを開始しますか？",
    confirmEnd: "ライブを終了しますか",
    prompt: "ライブ配信のアドレスを入力してください。",
    form: {
      img: "ライブイメージ",
      category: "ライブカテゴリー",

      "1": "映像ライブ",
      "2": "チャットライブ",

      user: "ライブモデレーター",
      title: "ライブタイトル",
      desc: "ライブ概要",
    },
    livePreview: "ライブプレビュー",
    action: {
      stop: "終了",
      delete: "削除",
      edit: "編集",
      share: "シェア",
      start: "開始",
      top: "ピン",
      cancelTop: "ピンを解除",
    },
    confirmTop: "ピンを解除しますか？",
    confirmCancelTop: "ピン留めしますか？",

    liveReplay: {
      title: "再生",
      delete: "削除",
    },
    stat: {
      title: "ライブデータ",
      read_count: "視聴人数",
      deliver_count: "拡散人数",
      duration: "ライブ放送時間（分）",
      user_count: "コメント人数",
      msg_count: "コメント数",
    },
  },

  invite: {
    title: "グループへの招待",
    desc: "グループに招待する",
    card: "招待状を送る",
    link: "グループへの招待リンクをコピー",
    tip1: "管理者は、メンバー管理から直接メンバーを追加することも可能です。",
    tip2: "招待特典を開始しました。友達をグループに招待し、デイリーボーナス抽選に使えるエナジーを獲得しましょう！",
    tip3: "あなたへの通報が多い場合は、報酬が没収されます。むやみな招待状の送信はお控えください。",
    my: {
      title: "招待履歴",
      reward: "招待報酬",
      people: "招待人数",
      noInvited: "招待履歴はありません",
      rule: "利用規定",
    },
    claim: {
      title: "友達をグループに招待すると、エナジー報酬プレゼント",
      btn: "招待する",
      count: "招待済み {count} 人"
    }
  },

  transfer: {
    title: "{name}トレード",
    price: "価格",
    pool: "プール",
    earn: "24時間年中無休のマーケットメイキング",
    amount: "24時間出来高",
    method: "トレード方法",
    noPrice: "現在価格がありません",

    order: "最大注文数 {amount} {symbol}",

    maker: "オートマーケットメイカー",
    taker: "{exchange}代理購入",

    Huobi: "Huobi",
    BigONE: "BigONE",
    Binance: "Binance",
    ExinSwap: "ExinSwap",
    MixSwap: "MixSwap",
    exchange: "取引所",
    sign: "個人間マルチシグネチャ取引",

    coin: "coin-coinトレード",
    otc: "fiat-coinトレード",

    auth: "Blue Shield Authentication",
    identity: "本名",

    payMethod: "支払い方式",
    bank: "銀行",
    alipay: "支付宝",
    wechatpay: "微信",

    category: "通貨",
    limit: "限度",
    in5minRate: "5分完了率",
    orderSuccessRank: "注文完了率",
    multisigOrderCount: "総トランザクション数",
  },

  reward: {
    title: "リワード",
    who: "送付先",
    amount: "数量",
    less: "最低額は1ドルです。",
    success: "リワードに成功しました。",
    isLiving: "ライブ中はリワードを行えません。ライブ終了後に実行ください。",
  },

  claim: {
    title: "デイリーボーナス",
    tag: "demo",
    receive: "賞品を受け取る",
    receiveSuccess: "受取成功。後ほど送金されます。",
    drew: "当選。",
    worth: "{prefix} ${value}",
    now: " ",
    you: " ",
    ticketCount: "回抽選可能",
    success: "ログインに成功しました。",
    ok: "閉じる",
    join: "コミュニティに参加",
    open: "コミュニティに参加",

    energy: {
      title: "エナジー",
      describe: "100エナジーで1回抽選",
      exchange: "今すぐ交換",
      success: "交換成功",
      checkin: {
        label: "ログイン",
        checked: "参加済み",
        count: "今週 {count}/7",
        describe: "ログイン毎にに10エナジー、1週間に5日ログインで50エナジーを獲得します",
      },
    },
    records: {
      title: "抽選記録",
      winning: "当選記録",
      energy: "エナジー記録",
      lottery: "抽選",
      power_lottery: "パワー抽選",
      power_claim: "デイリーログイン",
      power_claim_extra: "週5回参加",
      power_invitation: "招待報酬",
    },
  },
  guess: {
    name: "価格を予想して{coin}を獲得",
    up: "アップ",
    down: "ダウン",
    flat: "フラット",
    sure: "確定",
    notsure: "考え直す",
    okay: "了承しました",
    goChoose: "遊びに行く",
    todayGuess: "{coin} 今日の価格を予想する",
    todyDesc:
      "今日 {time} UTC+8 {coin} の価格は ${usd}です(参照:Coingecko.com)。明日の{time} UTC+8 の価格を予想してください:",

    choose: {
      tip: "ヒント",
      info: "予測されるトレンドがまだありません。選択して確定してください。",
    },
    confirm: {
      tip: "ヒント",
      info: "確定後は変更できません。価格予想選択を確定しますか？",
    },
    success: {
      tip: "おめでとうございます",
      info: "本日の予測には参加済みです。結果は、明日 {start} UTC+8 、 {coin} 価格が判明した時点で発表されます。",
    },
    notstart: {
      tip: "開始前です",
      info: "本日の予測はまだ始まっていません。{start} - {end} UTC+8 に参加ください。",
    },
    missing: {
      tip: "本日分終了",
      info: "本日の予測期間は終了致しました。明日{start} - {end} UTC+8 に参加ください。",
    },
    end: {
      tip: "コンテスト終了",
      info: "コンテストは終了しました。「参加履歴」ページで参加履歴をご確認ください。",
    },
    records: {
      name: "参加履歴",
      history: "{coin} 価格予想記録",
      consecutiveplay: "連続参加しています",
      condition: "3日連続で参加して、賞金のシェアに参加しましょう",
      play: "参加済み",
      guess: "予想する",
      up: "アップ",
      down: "ダウン",
      flat: "フラット",
      day: "日",
      vip: "VIPユーザーになる ",
      playresult: "参加結果:",
      date: "日付",
      result: "結果",
      win: "勝ち",
      lose: "負け",
      pending: "結果待ち",
      notplay: "不参加",
      notstart: "開始前です",
    },
  },
  trading: {
    rule: "トレード規定",
    time: "イベント時間",
    reward: "トレード報酬",
    auth: "参加を承認",
    viewRank: "視聴ランキング",
    modalDesc: "MixSwap または 4swap を介して、USDT、BTC、ETH を {symbol} とトレードすることで、トレードコンテストに参加できます。",
    mixSwap: "MixSwap でトレードする",
    swap: "4swap でトレードする",
    rank: "トレードランキング",
    ranked1: "第 {i} 位",
    amount: "あなたの取引量は {amount} {symbol}です。",
    ranked2: "第 {i} 位",
    noRank: "ランキングTOP10にランクインしていません。",
  },
  modal: {
    check: "支払い結果の確認",
    loading: "読み込み中",
  },

  action: {
    tips: "確認",
    cancel: "取消",
    save: "保存",
    confirm: "確認",
    submit: "提出",
    continue: "続ける",
    know: "わかりました",
    open: "オープン",
  },

  success: {
    copy: "コピーに成功しました",
    send: "送信に成功しました",
    operator: "操作に成功しました",
    save: "保存に成功しました",
    modify: "編集に成功しました",
  },
  error: {
    people: "人数が異なります",
    amount: "金額が異なります",
    mixin: "Linksからアクセスしてください",
    empty: "空欄不可",
    modify: "編集に失敗しました",
  },
}
const i18n = {}
getI18n(_i18n, i18n)
export default i18n
