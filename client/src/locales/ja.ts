import { getI18n } from "@/locales/tools"

const _i18n = {
  site: {
    title: "スーパーコミュニティ",
  },
  pre: {
    create: {
      title: "オープンコミュニティ",
      desc: "コミュニティ・アシスタントは、新しいグループの自動作成、会員登録の定期的なチェック、グループへのラッキーコインの送付、アナウンスなどを行うことができます。",
      button: "0.01XINを支払ってオープンコミュニティを作成する",
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
      "コミュニティチャットの受信を停止しても、重要なプロジェクトのアナウンスはすべて掲載されます。",
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
        "コミュニティチャットの受信を停止しても、重要なプロジェクトのアナウンスはすべて掲載されます。<br />以下の数字を順番に入力して、操作を確定してください。",
    },
    exited: "コミュニティから退会しました",
    exitedDesc:
      "コミュニティから退会しました。また、あなたのアカウントに関するデータはすべて削除されました。右上をクリックしてコミュニティページを閉じてください。",
  },

  manager: {
    setting: "設定",
    base: "基本設定",
    description: "プロフィール",
    welcome: "ようこそ、グループへ",
    high: "詳細設定",

    helloTips: "ウェルカムメッセージは、グループに新しく参加した人だけが見ることができ、他のグループのメンバーは見ることができません。",
  },
  broadcast: {
    a: "アナウンス",
    title: "アナウンスの設定",
    holder: "ユーザーにアナウンスしたい内容を記入してください",
    recall: "キャンセル",
    confirmRecall: "本当にキャンセルしますか？",
    recallSuccess: "キャンセルしました",
    status0: "送信中",
    status1: "送信済み",
    status2: "キャンセル処理中",
    status3: "キャンセル済み",
    checkNumber: "数字が一致していることをご確認してください",
    sent: "グループへのアナウンス",
    input: "大量のアナウンス内容を送信するために、上記の数字を入力してください",
    fill: "最初にアナウンスしたい内容を記入してください",
    send: "送信",
  },
  stat: {
    title: "統計情報",
    totalUser: "総ユーザー数",
    highUser: "最少ユーザー数",
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
    hour: "直近{n}時間の活動状況",
    day: "直近{n}日間の活動状況",
    month: "直近{n}ヶ月の活動状況",
    year: "直近{n}年の活動状況",
    action: {
      set: "{c}に設定します",
      cancel: "{c}をキャンセルします",
      confirmSet: "{full_name}({identity_number})が{c}に設定されていることを確認してください。",
      confirmCancel: "{full_name}({identity_number})を{c}にすることをキャンセルしますか？",
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
      desc: "ユーザーは1時間ミュートされます。ミュート後もメッセージの受信やラッキーコインの取得には影響しません。",
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

    center: "メンバーシップ",
    level0: "無料メンバー",
    level0Desc:
      "1-全てのチャットを受信することができます,1-ラッキーコインに参加することができます,1-チャットに参加するためにメッセージを送信することができます,1-1分間に5〜20メッセージを送信する事ができます,1-シニアメンバーは写真、ビデオ、他の種類のメッセージを送ることができます。",
    level0Sub:
      "定期的なウォレットチェックの際のウォレット内ポジションに応じて、会費無料で、ポジションジュニアメンバーまたはポジションシニアメンバーに自動登録されます。ウォレット残高が{lamount} {symbol}以上の場合は無料ジュニアメンバー、{hamount} {symbol}以上の場合は無料シニアメンバーになります。",
    level1: "メンバーシップ無し",
    level1Desc: "1-全てのチャットを受信することができます,1-ラッキーコインに参加することができます,0-チャットに参加するためにメッセージを送信することができます",
    level2: "ジュニアメンバー",
    level2Auth: "ポジションジュニアメンバー",
    level2Pay: "有料ジュニアメンバー",
    level2Desc:
      "1-全てのチャットを受信することができます,1-ラッキーコインに参加することができます,1-チャットに参加するためにメッセージを送信することができます,1-1分間に5〜10メッセージを送信する,1-テキストなど3種類のメッセージを送信することができます。",
    level2Sub: "テキストを含む3種類のメッセージを送信でき、1分間に5〜10通のメッセージを送信できます。",
    level5: "シニアメンバー",
    level5Auth: "ポジションシニアメンバー",
    level5Pay: "有料シニアメンバー",
    level5Desc:
      "1-全てのチャットを受信することができます,1-ラッキーコインに参加することができます,1-チャットに参加するためにメッセージを送信することができます,1-1分間に20メッセージを送信することができます,1-テキストを含む9種類のメッセージを送信することができます。",
    level5Sub: "テキストを含む9種類のメッセージを送信でき、1分間に10〜20通の送信が可能です。",

    upgrade: "メンバーシップのアップグレード",
    levelPay:
      "{amount} {symbol}を支払って1年分の{level}を取得でき、テキストなど{category}のメッセージを送信でき、1分間に{min}～{max}通のメッセージの送信が可能です。",
    checkPaid: "支払確認",
    authTips:
      "ウォレット内ポジションによるフリーメンバーシップは、定期的にお客様の資産がメンバーシップの要件を満たしているかを確認します。詳細な手順については、ドキュメントをご覧ください。：<a href='https://w3c.group/c/1628159023237756'>https://w3c.group/c/1628159023237756</a>",
    forFree: "無料で利用可能",
    forPay: "{amount} {symbol}の支払いを確認しました",

    cancel: "メンバーシップの解除",
    cancelDesc:
      "下記の「メンバーシップの解除」をクリックするとメンバーシップ資格を失い、コミュニティロボがあなたの資産情報を読み取ることができなくなります。",

    expire: "メンバーシップ資格は{date}まで有効です。期限時に更新してください。",
    failed: "メンバーシップの有効化に失敗しました。",
    failedDesc:
      "あなたはメンバーシップ有効化の要件を満たしていません。資産を読み取りが許可されていることを確認してください。 あなたの資産がExinOneの流動性プールにある場合、ExinOneロボを開いて資産ページに切り替え、上部の設定アイコンをクリックして資産の認証を許可してください。",
  },
  advance: {
    title: "詳細設定",
    mute: "グループ全体をミュート",
    muteConfirm: "グループ全体をミュート{action}しますか{tips}？",
    muteTips: "（管理者とゲストは対象外です）",
    open: "開始",
    close: "終了",
    newMember: "グループ参加通知",
    newMemberConfirm: "グループ参加の通知{action}を行いますか？",
    sliderConfirm: "スライドして操作を確認する",
    proxy: "メッセージのやり取り禁止",
    proxyConfirm: "メッセージのやり取り{action}を禁止しますか？",
    proxyTips: "（管理者は対象外です）",
    msgAuth: "メッセージの許可",
    member: {
      1: "メンバー外",
      2: "ジュニアメンバー",
      5: "シニアメンバー",
      tips: "{status}が1分間に送信できるメッセージの最大数は{count}です。"
    },
    plain_text: "文字",
    plain_sticker: "ステッカー",
    plain_image: "画像",
    plain_video: "動画",
    lucky_coin: "ラッキーコイン",
    plain_post: "投稿",
    plain_live: "ライブ配信",
    plain_contact: "ロボの連絡先",
    plain_transcript: "チャット記録",
    plain_data: "ファイル",
    url: "リンク",
    app_card: "カード",
  },
  join: {
    title: "コミュニティを探す",
    received: "受取成功",
    open: "Mixinで開く",

    main: {
      join: "グループへ参加",
      joinTips: "【注意】Mixinはプロジェクトが発行する暗号資産を推奨したり、プロジェクトの安全性を保証したりするものではありません。",

      appointBtn: "予約",
      appointedBtn: "予約済み",
      appointedTips: "連絡先の追加",

      member: "メンバー",

      receiveBtn: "エアドロップの受け取り",

      receivedBtn: "すでにエアドロップは受け取り済みです",
      receivedTips: "エアドロップに参加する",

      noAccess: "参加資格がありません",

      appointOver: "エアドロップは終了しました",
    },

    modal: {
      auth: "認証に失敗しました",
      authDesc: "資産へのアクセス許可に同意してください。データはメンバーシップ情報の検出のみに使用されます。",
      authBtn: "再認証",

      forbid: "グループへの参加を禁止する",
      forbidDesc1: "24時間以内にグループに参加できない場合は、管理者に連絡するか、24時間後に再参加してください。",
      forbidDesc2: "あなたはグループから退会させられました。グループに再参加するには、管理者に連絡してください。",
      forbidBtn: "わかりました",

      shares: "メンバーシップ情報の検出に失敗しました",
      sharesBtn: "再検出",
      sharesFail: "検出不可",
      sharesTips: "今すぐ購入する",
      sharesCheck: "以下の保有資産のいずれかを満たしていることをご確認ください。",
      sharesCheck1: "以上",
      sharesCheck2:
        "メンバーシップ情報の検出は、Mixinウォレット、Exinの流動性プール、活期宝、省心投和 Fox 活期理财、定期理财、可盈池に対応しています。",

      appoint: "予約が完了しました",
      appointDesc:
        "ご予約ありがとうございます。 エアドロップの通知を即時に受け取るために、現在のロボの連絡先を追加し、通知許可をオンにしてください。",
      appointBtn: "連絡先の追加",
      appointTips: "右上をタップしてロボの通知をオフにする",
      receive: "エアドロップ報酬",
      receiveDesc:
        "MobileCoinエアドロップへの参加資格取得、おめでとうございます。 MobileCoinのサポートに感謝し、エアドロップを受け取り、グループに参加することを歓迎します。",
      receiveBtn: "MOBを受け取る",
      receivedDesc:
        "{comment}に参加された方へのエアドロップ報酬です！いつもMobileCoinとMixinを応援して頂き、ありがとうございます。",
      receivedBtn: "受取済み",
    },

    code: {
      invite: "Links でグループに参加する",
      download: "Links をダウンロードする",
      href: "https://getlinks.jp/",
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

    people_count: "コミュニティ人数",
    week: "今週",
    trade: "トレード",
    invite: "招待",
    findGroup: "コミュニティを探す",
    findBot: "ロボを探す",
    activity: "イベント情報",
    redPacket: "ラッキーコイン",
    reward: "リワード",
    claim: "サインイン抽選",
    open: "オープンチャット",
    article: "インフォメーション",
    more: "その他のイベント",
    noActive: "イベントはありません。",
    noNews: "情報なし",
    notStart: "イベントは開始されていません。",
    isEnd: "イベントは終了しました。",

    joinSuccess: "参加に成功しました。",
    enterChat: "チャットに参加する",
    enterHome: "コミュニティホームへ",
  },

  news: {
    all: "すべて",
    replay: "アーカイブス",
    broadcast: "アナウンス",
    sendBroadcast: "アナウンスを送信",
    sendLive: "ライブの追加",
    live: "ライブ",
    confirmStart: "ライブを開始しますか？",
    confirmEnd: "ライブを終了しますか",
    prompt: "ライブ配信のアドレスを入力してください。",
    form: {
      img: "ライブマップ",
      category: "ライブカテゴリー",

      "1": "映像ライブ",
      "2": "画像+音声ライブ",

      user: "ライブモデレーター",
      title: "ライブタイトル",
      desc: "ライブ概要",
    },
    livePreview: "ライブプレビュー",
    action: {
      stop: "ライブストップ",
      delete: "削除",
      edit: "トレーラー編集",
      share: "トレーラーシェア",
      start: "ライブスタート",
      top: "トップ",
      cancelTop: "トップをキャンセル",
    },
    confirmTop: "トップを確定しますか？",
    confirmCancelTop: "トップのキャンセルを確定しますか？",

    liveReplay: {
      title: "ライブプレイバック",
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
    link: "グループに参加するにはリンクをコピーしてください。",
    tip1: "管理者は、メンバー管理から直接メンバーを追加することも可能です。",
    tip2: "グループ特典への招待を開始しました！友達を誘ってコミュニティに参加し、サインアップ抽選会に参加してエナジー報酬を受け取りましょう！",
    tip3: "他人への嫌がらせをしないでください！あなたの通報が多い場合は、すべての報酬が即座にキャンセルされます！",
    my: {
      title: "自分の招待状",
      reward: "招待報酬",
      people: "招待人数",
      noInvited: "招待はありません",
      rule: "利用規定",
    },
    claim: {
      title: "友人を任意のコミュニティに招待し、エナジー報酬を得る。",
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
    taker: "？{exchange}代理購入",

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
    who: "送り先",
    amount: "数量",
    less: "最低額は1ドルです。",
    success: "リワードに成功しました。",
    isLiving: "音声ライブ流派報酬を受け取ることができません。ライブ終了後にお願いします。",
  },

  claim: {
    title: "ボーナス抽選",
    tag: "試運転",
    receive: "賞品を受け取る",
    receiveSuccess: "受取成功。後ほど送信されます。",
    drew: "抽選済み",
    worth: "{prefix}価値 $ {value}",
    now: "今すぐ",
    you: "あなたは",
    ticketCount: "回抽選に参加できます",
    success: "チェックアップに成功しました。",
    ok: "OK",
    join: "コミュニティに参加する。",
    open: "コミュニティに参加",

    energy: {
      title: "エネルギー",
      describe: "100エナジーで1回抽選",
      exchange: "今すぐ交換",
      success: "交換成功",
      checkin: {
        label: "チェックアップ",
        checked: "チェックアップ済み",
        count: "今週{count}/7",
        describe: "チェックアップすると10エナジー、1週間内に5日間チェックアップで50エナジー追加",
      },
    },
    records: {
      title: "抽選記録",
      winning: "当選記録",
      energy: "エナジー記録",
      lottery: "抽選",
      power_lottery: "パワー抽選",
      power_claim: "デイリーチェックアップ",
      power_claim_extra: "今週5回チェックアップ",
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
    tips: "ヒント",
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
    mixin: "Links内から開いてください",
    empty: "空欄不可",
    modify: "編集に失敗しました",
  },
}
const i18n = {}
getI18n(_i18n, i18n)
export default i18n
