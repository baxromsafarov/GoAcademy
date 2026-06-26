const fs = require("fs")
const path = require("path")
const PptxGenJS = require("pptxgenjs")

// ---- dark theme matching the app ----
const BG = "13151D" // app dark background
const SURFACE = "1E2230" // card surface
const PRIMARY = "6E8BFF" // app dark primary (indigo)
const INK = "ECEDEF" // foreground
const MUTED = "9AA4B2"
const FONT = "Meiryo"

const SHOTS = path.join(__dirname, "shots")
const img = (f) => path.join(SHOTS, f)
const IMG_RATIO = 1380 / 860 // capture viewport

const pptx = new PptxGenJS()
pptx.defineLayout({ name: "W", width: 13.333, height: 7.5 })
pptx.layout = "W"
pptx.author = "GoAcademy"
pptx.title = "GoAcademy"

// ---------- Title slide ----------
const t = pptx.addSlide()
t.background = { color: BG }
t.addShape(pptx.ShapeType.rect, { x: 0, y: 3.05, w: 13.333, h: 0.06, fill: { color: PRIMARY } })
t.addText("GoAcademy", {
  x: 0.9, y: 1.5, w: 11.5, h: 1.3, fontFace: FONT, fontSize: 60, bold: true, color: PRIMARY,
})
t.addText("Go言語を初心者から実務レベルまで学べる学習プラットフォーム", {
  x: 0.9, y: 3.25, w: 11.5, h: 0.7, fontFace: FONT, fontSize: 22, color: INK,
})
t.addText("ブラウザでコードを動かしながら学ぶ｜動画・記事・クイズ・課題・プロジェクト", {
  x: 0.9, y: 4.0, w: 11.5, h: 0.6, fontFace: FONT, fontSize: 16, color: MUTED,
})
t.addText("4言語対応：日本語・英語・ロシア語・ウズベク語", {
  x: 0.9, y: 4.5, w: 11.5, h: 0.6, fontFace: FONT, fontSize: 16, color: MUTED,
})
t.addText("発表者：＿＿＿＿＿＿", {
  x: 0.9, y: 6.4, w: 6, h: 0.5, fontFace: FONT, fontSize: 14, color: MUTED,
})

// ---------- content slides ----------
const slides = [
  {
    n: "02", title: "課題",
    bullets: [
      "Go言語の学習教材はあちこちに分散していて、体系的に学びにくい",
      "学んだ知識をその場で試せる環境が少ない",
      "一人だと進捗ややる気を保ちにくい",
      "言語の壁：母国語で学べる教材が限られている",
    ],
  },
  {
    n: "03", title: "GoAcademyとは", img: "01-dashboard.png",
    bullets: [
      "動画・記事・クイズ・課題を一つにまとめた学習プラットフォーム",
      "ブラウザ上でGoのコードをそのまま実行できる",
      "学習の進捗を記録し、ゲーム感覚で続けられる",
      "初心者から実務レベルまで、順序立てて学べる",
    ],
  },
  {
    n: "04", title: "主な機能", img: "02-videos.png",
    bullets: [
      "動画、記事、クイズ、アルゴリズム課題",
      "学習トラック（コース）とミニプロジェクト",
      "チートシート、用語集",
      "ブックマーク（保存）と学習メモ",
    ],
  },
  {
    n: "05", title: "コードサンドボックス", img: "04-sandbox.png",
    bullets: [
      "ブラウザからGoのコードを実行できる",
      "安全なDockerコンテナで隔離して実行",
      "記事の中のコードもその場で実行可能",
      "課題は自動採点（オンラインジャッジ）",
    ],
  },
  {
    n: "06", title: "学習体験とゲーミフィケーション", img: "05-track.png",
    bullets: [
      "コース → レッスンの順序立てた学習の流れ",
      "経験値（XP）・レベル・連続学習日数",
      "バッジとランキングでモチベーション維持",
      "ダッシュボードで進捗を一目で確認",
    ],
  },
  {
    n: "07", title: "多言語対応", img: "03-article.png",
    bullets: [
      "日本語・英語・ロシア語・ウズベク語に対応",
      "インターフェースもコンテンツも言語ごとに用意",
      "日本語の学習者向けに日本語のGo動画を収録",
      "言語を切り替えても設定が保存される",
    ],
  },
  {
    n: "08", title: "技術スタック",
    bullets: [
      "バックエンド：Go（chi、pgx、sqlc）",
      "フロントエンド：React、TypeScript、Vite、Tailwind CSS",
      "データベース：PostgreSQL",
      "認証：JWT｜コード実行：Dockerによる安全なサンドボックス",
    ],
  },
  {
    n: "09", title: "管理機能（管理者向け）", img: "07-admin.png",
    bullets: [
      "管理画面で全コンテンツを作成・編集・削除",
      "コンテンツの表示／非表示を切り替え",
      "ユーザー管理と権限の設定",
      "検索・フィルター・ページネーション",
    ],
  },
  {
    n: "10", title: "まとめと今後",
    bullets: [
      "Goを学ぶための一体型プラットフォーム：学習・実践・進捗管理を一つに",
      "ブラウザだけで「読む → 動かす → 試す」が完結",
      "今後：コンテンツの拡充、モバイル対応、AIによる学習サポート",
      "ご清聴ありがとうございました",
    ],
  },
]

for (const s of slides) {
  const slide = pptx.addSlide()
  slide.background = { color: BG }
  // header
  slide.addShape(pptx.ShapeType.rect, { x: 0.55, y: 0.55, w: 0.13, h: 0.62, fill: { color: PRIMARY } })
  slide.addText(s.title, {
    x: 0.85, y: 0.5, w: 11.8, h: 0.75, fontFace: FONT, fontSize: 30, bold: true, color: INK, valign: "middle",
  })
  slide.addShape(pptx.ShapeType.rect, { x: 0.55, y: 1.4, w: 12.25, h: 0.02, fill: { color: "2A3142" } })

  const hasImg = s.img && fs.existsSync(img(s.img))
  const bx = 0.7
  const bw = hasImg ? 5.4 : 11.9
  slide.addText(
    s.bullets.map((b) => ({
      text: b,
      options: { bullet: { code: "2022", indent: 20 }, paraSpaceAfter: 16, color: INK },
    })),
    { x: bx, y: 1.85, w: bw, h: 5.0, fontFace: FONT, fontSize: hasImg ? 18 : 22, color: INK, valign: "top", lineSpacingMultiple: 1.2 },
  )

  if (hasImg) {
    const w = 6.7
    const h = w / IMG_RATIO
    const x = 6.25
    const y = 1.7 + (5.2 - h) / 2
    // subtle frame
    slide.addShape(pptx.ShapeType.rect, { x: x - 0.06, y: y - 0.06, w: w + 0.12, h: h + 0.12, fill: { color: SURFACE }, line: { color: "39415A", width: 1 } })
    slide.addImage({ path: img(s.img), x, y, w, h })
  }

  slide.addText("GoAcademy", { x: 0.7, y: 6.98, w: 6, h: 0.35, fontFace: FONT, fontSize: 11, color: MUTED })
  slide.addText(s.n, { x: 12.1, y: 6.98, w: 0.7, h: 0.35, fontFace: FONT, fontSize: 11, color: MUTED, align: "right" })
}

pptx
  .writeFile({ fileName: "GoAcademy_JP.pptx" })
  .then((f) => console.log("wrote " + f))
  .catch((e) => { console.error(e); process.exit(1) })
