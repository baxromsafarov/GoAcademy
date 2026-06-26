const fs = require("fs")
const path = require("path")
const puppeteer = require("puppeteer")

const CHROME = "C:/Program Files/Google/Chrome/Application/chrome.exe"
const BASE = "http://localhost:5173"
const EMAIL = "demo-ppt@example.com"
const PASS = "Passw0rd!23"
const TRACK = "867fe28b-d84e-4ec3-a093-820723621347"
const PROBLEM = "go-min2-ja"
const ARTICLE = "go-select-ja"

const OUT = path.join(__dirname, "shots")
fs.mkdirSync(OUT, { recursive: true })

const sleep = (ms) => new Promise((r) => setTimeout(r, ms))

async function main() {
  const browser = await puppeteer.launch({
    executablePath: CHROME,
    headless: "new",
    args: ["--no-sandbox", "--hide-scrollbars"],
    defaultViewport: { width: 1380, height: 860, deviceScaleFactor: 2 },
  })
  const page = await browser.newPage()
  // Dark theme + Japanese UI before the app boots.
  await page.evaluateOnNewDocument(() => {
    localStorage.setItem("theme", "dark")
    localStorage.setItem("lang", "ja")
  })

  // --- log in via the UI form ---
  await page.goto(`${BASE}/login`, { waitUntil: "networkidle0" })
  await page.waitForSelector("#email")
  await page.type("#email", EMAIL)
  await page.type("#password", PASS)
  await Promise.all([
    page.click('button[type="submit"]'),
    page.waitForFunction(() => location.pathname === "/", { timeout: 20000 }).catch(() => {}),
  ])
  await sleep(2500) // let dashboard data + charts render

  const shots = [
    ["01-dashboard", "/"],
    ["02-videos", "/videos"],
    ["03-article", `/articles/${ARTICLE}`],
    ["04-sandbox", "/sandbox"],
    ["05-track", `/tracks/${TRACK}`],
    ["06-problem", `/problems/${PROBLEM}`],
    ["07-admin", "/admin"],
  ]

  for (const [name, route] of shots) {
    try {
      // Client-side (SPA) navigation — no full reload, so the in-memory session
      // (and dark theme) survive between screenshots.
      await page.evaluate((r) => {
        window.history.pushState({}, "", r)
        window.dispatchEvent(new PopStateEvent("popstate"))
      }, route)
      await sleep(2200)
      if (name === "03-article") {
        // reveal the run/open toolbar that appears on hover
        const box = await page.$("pre")
        if (box) {
          const b = await box.boundingBox()
          if (b) await page.mouse.move(b.x + b.width / 2, b.y + 20)
          await sleep(500)
        }
      }
      await page.screenshot({ path: path.join(OUT, `${name}.png`) })
      console.log("shot", name)
    } catch (e) {
      console.error("FAILED", name, e.message)
    }
  }

  await browser.close()
}

main().catch((e) => {
  console.error(e)
  process.exit(1)
})
