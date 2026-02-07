#!/usr/bin/env node
'use strict';

const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const process = require('node:process');
const readline = require('node:readline/promises');

function maybeSetPlaywrightHostPlatformOverride() {
  // In some restricted environments, `os.cpus()` can be empty even on Apple Silicon.
  // Playwright uses that to detect arm64 Macs; when it fails, it may look for mac-x64 browser builds.
  if (process.platform !== 'darwin') return;
  if (process.arch !== 'arm64') return;
  if (process.env.PLAYWRIGHT_HOST_PLATFORM_OVERRIDE) return;

  const cpus = os.cpus();
  if (!Array.isArray(cpus) || cpus.length > 0) return;

  // Mirror Playwright's darwin mapping: Darwin major 24 => mac15.
  const ver0 = parseInt(String(os.release()).split('.')[0], 10);
  if (!Number.isFinite(ver0)) return;

  const LAST_STABLE_MACOS_MAJOR_VERSION = 15;
  const macMajor = Math.min(ver0 - 9, LAST_STABLE_MACOS_MAJOR_VERSION);
  if (macMajor < 10) return;

  process.env.PLAYWRIGHT_HOST_PLATFORM_OVERRIDE = `mac${macMajor}-arm64`;
  console.log(`Note: set PLAYWRIGHT_HOST_PLATFORM_OVERRIDE=${process.env.PLAYWRIGHT_HOST_PLATFORM_OVERRIDE} (restricted environment detected).`);
}

function usage() {
  // Keep this terse; README has the full docs.
  console.log(`Usage:
  node scripts/auth-capture.js [--config auth.config.json] [--appName <name>] [--baseURL <url>] [--loginURL <urlOrPath>] [--overwrite]

Examples:
  node scripts/auth-capture.js
  node scripts/auth-capture.js --appName staging --baseURL http://localhost:4500 --loginURL /login --overwrite
`);
}

function parseArgs(argv) {
  const out = {};
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '-h' || a === '--help') out.help = true;
    else if (a.startsWith('--') && a.includes('=')) {
      const [k, ...rest] = a.slice(2).split('=');
      out[k] = rest.join('=');
    } else if (a.startsWith('--')) {
      const k = a.slice(2);
      const v = argv[i + 1];
      if (!v || v.startsWith('--')) out[k] = true;
      else {
        out[k] = v;
        i++;
      }
    } else {
      out._ = out._ || [];
      out._.push(a);
    }
  }
  return out;
}

function safeFilenameSegment(s) {
  return String(s).trim().replace(/[^a-zA-Z0-9._-]/g, '_');
}

function readJSONIfExists(p) {
  if (!fs.existsSync(p)) return null;
  const raw = fs.readFileSync(p, 'utf8');
  return JSON.parse(raw);
}

function mustURL(raw, field) {
  try {
    return new URL(raw).toString();
  } catch (e) {
    throw new Error(`Invalid ${field}: ${JSON.stringify(raw)}`);
  }
}

function resolveURL(raw, baseURL) {
  // Allow relative paths like "/login" (resolved against baseURL).
  try {
    return new URL(raw, baseURL).toString();
  } catch (e) {
    throw new Error(`Invalid loginURL: ${JSON.stringify(raw)}`);
  }
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    usage();
    process.exit(0);
  }

  const configPath = path.resolve(process.cwd(), args.config || 'auth.config.json');
  const fileCfg = readJSONIfExists(configPath) || {};

  const appName = (args.appName || fileCfg.appName || '').trim();
  const baseURLRaw = (args.baseURL || fileCfg.baseURL || '').trim();
  const loginURLRaw = (args.loginURL ?? fileCfg.loginURL ?? '').trim();

  if (!appName) throw new Error(`Missing appName. Set it in ${path.basename(configPath)} or pass --appName.`);
  if (!baseURLRaw) throw new Error(`Missing baseURL. Set it in ${path.basename(configPath)} or pass --baseURL.`);

  const baseURL = mustURL(baseURLRaw, 'baseURL');
  const targetURL = loginURLRaw ? resolveURL(loginURLRaw, baseURL) : baseURL;

  const safeApp = safeFilenameSegment(appName);
  const authDir = path.resolve(process.cwd(), '.auth');
  const outPath = path.join(authDir, `${safeApp}.json`);
  fs.mkdirSync(authDir, { recursive: true });
  if (safeApp !== appName) {
    console.log(`Note: appName ${JSON.stringify(appName)} will be saved as ${JSON.stringify(safeApp)} for the filename.`);
  }

  if (fs.existsSync(outPath) && !args.overwrite) {
    const rl0 = readline.createInterface({ input: process.stdin, output: process.stdout });
    const ans = (await rl0.question(
      `Auth state already exists at ${path.relative(process.cwd(), outPath)}.\nOverwrite it? (y/N) `
    )).trim().toLowerCase();
    rl0.close();
    if (ans !== 'y' && ans !== 'yes') {
      console.log('Keeping existing auth state (no changes made).');
      process.exit(0);
    }
  }

  let playwright;
  try {
    // This repo gitignores package.json/node_modules, so users typically install locally and keep it untracked:
    //   npm i -D playwright
    maybeSetPlaywrightHostPlatformOverride();
    playwright = require('playwright');
  } catch (e) {
    console.error(`Playwright is not installed (can't resolve "playwright").\n`);
    console.error(`Install it repo-locally (untracked in this repo) and try again:\n`);
    console.error(`  npm i -D playwright`);
    console.error(`  npx playwright install chromium\n`);
    process.exit(1);
  }

  const { chromium } = playwright;

  const browser = await chromium.launch({ headless: false });
  const context = await browser.newContext({
    baseURL,
    ignoreHTTPSErrors: true
  });
  const page = await context.newPage();

  const closeAndExit = async (code) => {
    try {
      await browser.close();
    } catch (_) {
      // ignore
    }
    process.exit(code);
  };

  process.on('SIGINT', () => {
    console.log('\nCaught SIGINT, closing browser...');
    void closeAndExit(130);
  });

  console.log(`Opening: ${targetURL}`);
  await page.goto(targetURL, { waitUntil: 'domcontentloaded' });

  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
  await rl.question('Log in in the opened browser, then press Enter here to save storageState...\n');
  rl.close();

  await context.storageState({ path: outPath });
  await browser.close();

  console.log(`Saved storageState to: ${path.relative(process.cwd(), outPath)}`);
}

main().catch((err) => {
  console.error(err?.stack || String(err));
  process.exit(1);
});
