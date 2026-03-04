<script>
  import { onMount, onDestroy } from "svelte";
  import { WindowSetSize, WindowSetMinSize, EventsOn, WindowSetPosition, WindowGetPosition } from "../wailsjs/runtime/runtime.js";
  import {
    GetConfig,
    GetCertInfo,
    GetRotatorStatus,
    GetUDPStatus,
    RotatorSetFollow,
    RotatorPark,
  } from "../wailsjs/go/main/App.js";
  import StatusTab from "./components/StatusTab.svelte";
  import ConfigTab from "./components/ConfigTab.svelte";
  import MiniMode from "./components/MiniMode.svelte";
  import CertBanner from "./components/CertBanner.svelte";

  // ── Window constants ───────────────────────────────────────────────────────
  const WIDTH       = 430;
  const FULL_HEIGHT = 620;
  const MINI_BASE   = 130;  // no rotator
  const MINI_ROT    = 185;  // + rotator (az/el + follow control)
  const MINI_MAP    = 295;  // + compact polar map

  // ── UI state ───────────────────────────────────────────────────────────────
  let miniMode  = localStorage.getItem("ui.miniMode") === "true";
  let activeTab = "status";
  let utcTime   = "";
  let clockInterval;

  // ── Shared runtime state ───────────────────────────────────────────────────
  let freqMHz       = "";
  let mode          = "";
  let split         = false;
  let freqTxMHz     = "";
  let modeTx        = "";
  let statusMsg     = "";
  let qsoResult     = null;
  let radioEnabled  = false;
  let rotatorEnabled = false;
  let rotConnected  = false;
  let rotAz         = 0;
  let rotEl         = 0;
  let rotFollow     = "off";
  let hfAz          = null;
  let satAz         = null;
  let satEl         = null;
  let minimapEnabled = false;
  let certInfo      = null;

  // ── Reactive mini-height ───────────────────────────────────────────────────
  $: miniHeight = rotatorEnabled
    ? (minimapEnabled ? MINI_MAP : MINI_ROT)
    : MINI_BASE;

  async function setWindowSizeAnchored(width, height) {
    const pos = await WindowGetPosition();
    WindowSetSize(width, height);
    WindowSetPosition(pos.x, pos.y);
  }

  // Re-apply window size when height changes while in mini mode
  $: if (miniMode) {
    WindowSetMinSize(WIDTH, miniHeight);
    setWindowSizeAnchored(WIDTH, miniHeight);
  }

  // ── Window helpers ─────────────────────────────────────────────────────────
  function enterMiniMode() {
    miniMode = true;
    localStorage.setItem("ui.miniMode", "true");
    WindowSetMinSize(WIDTH, miniHeight);
    setWindowSizeAnchored(WIDTH, miniHeight);
  }

  function exitMiniMode() {
    miniMode = false;
    localStorage.setItem("ui.miniMode", "false");
    WindowSetMinSize(WIDTH, FULL_HEIGHT);
    setWindowSizeAnchored(WIDTH, FULL_HEIGHT);
  }

  // ── Rotator actions ────────────────────────────────────────────────────────
  async function setFollow(followMode) {
    rotFollow = followMode;
    await RotatorSetFollow(followMode);
  }

  async function park() {
    rotFollow = "off";
    await RotatorPark();
  }

  // ── Clock ──────────────────────────────────────────────────────────────────
  function updateClock() {
    const now = new Date();
    utcTime = now.toUTCString().split(" ").slice(4, 5)[0] + " UTC";
  }

  // ── Lifecycle ──────────────────────────────────────────────────────────────
  let offRadio, offQso, offStatus, offRotPos, offRotStatus, offRotBearing,
      offProfile, offRadioEnabled, offRotEnabled, offAdvanced, offCert;

  onMount(async () => {
    updateClock();
    clockInterval = setInterval(updateClock, 1000);

    offRadio = EventsOn("radio:status", (data) => {
      if (data && data.freqMHz !== undefined) {
        freqMHz   = Number(data.freqMHz).toFixed(5);
        mode      = data.mode || "";
        split     = data.split || false;
        freqTxMHz = data.split ? Number(data.freqTxMHz).toFixed(5) : "";
        modeTx    = data.split ? (data.modeTx || "") : "";
      }
    });
    offQso = EventsOn("qso:result", (data) => {
      qsoResult = data;
      setTimeout(() => { qsoResult = null; }, 30000);
    });
    offStatus = EventsOn("status:message", (msg) => { statusMsg = msg; });
    offRotPos = EventsOn("rotator:position", (data) => {
      if (data) { rotAz = data.az; rotEl = data.el; }
    });
    offRotStatus = EventsOn("rotator:status", (connected) => { rotConnected = connected; });
    offRotBearing = EventsOn("rotator:bearing", (data) => {
      if (!data) return;
      if (data.type === "hf") { hfAz = data.az; }
      if (data.type === "sat") { satAz = data.az; satEl = data.el; }
    });
    offRotEnabled = EventsOn("rotator:enabled", (enabled) => {
      rotatorEnabled = enabled;
      if (!enabled) rotConnected = false;
    });
    offRadioEnabled = EventsOn("radio:enabled", (enabled) => { radioEnabled = enabled; });
    offProfile = EventsOn("profile:switched", (data) => {
      rotatorEnabled = data?.rotatorEnabled || false;
      radioEnabled   = data?.radioEnabled   || false;
      hfAz = null; satAz = null; satEl = null;
      rotFollow = "off"; rotConnected = false;
    });
    offAdvanced = EventsOn("advanced:changed", (data) => {
      minimapEnabled = data?.minimapEnabled ?? false;
    });
    offCert = EventsOn("cert:install_needed", (data) => { certInfo = data; });

    // Load initial state
    const cfg = await GetConfig();
    const p = cfg.profiles?.[cfg.profile];
    radioEnabled   = p?.flrig_ena || p?.hamlib_ena || false;
    rotatorEnabled = p?.rotator_enabled || false;
    if (rotatorEnabled) {
      const s = await GetRotatorStatus();
      rotConnected = s.connected;
      rotAz = s.az; rotEl = s.el;
      rotFollow = s.followMode || "off";
    }

    const adv = await GetUDPStatus();
    minimapEnabled = adv.minimapEnabled;

    const ci = await GetCertInfo();
    if (!ci.isInstalled) certInfo = ci;

    // Apply correct window size on startup
    if (miniMode) {
      WindowSetMinSize(WIDTH, miniHeight);
      WindowSetSize(WIDTH, miniHeight);
    }
  });

  onDestroy(() => {
    clearInterval(clockInterval);
    if (offRadio)      offRadio();
    if (offQso)        offQso();
    if (offStatus)     offStatus();
    if (offRotPos)     offRotPos();
    if (offRotStatus)  offRotStatus();
    if (offRotBearing) offRotBearing();
    if (offProfile)    offProfile();
    if (offRotEnabled) offRotEnabled();
    if (offRadioEnabled) offRadioEnabled();
    if (offAdvanced)   offAdvanced();
    if (offCert)       offCert();
  });
</script>

<div class="flex flex-col h-screen">
  {#if miniMode}
    <!-- ── MINI MODE ─────────────────────────────────────────────────────── -->
    <MiniMode
      {utcTime}
      {freqMHz} {mode} {split} {freqTxMHz} {modeTx} {qsoResult}
      {rotatorEnabled} {minimapEnabled}
      {rotConnected} {rotAz} {rotEl} {rotFollow}
      {hfAz} {satAz} {satEl}
      on:expand={exitMiniMode}
      on:follow={(e) => setFollow(e.detail)}
    />
  {:else}
    <!-- ── FULL MODE ──────────────────────────────────────────────────────── -->
    <header class="bg-surface-header flex items-center justify-between px-3 h-10 flex-shrink-0 border-b border-stroke-subtle">
      <div class="flex items-center gap-1 bg-surface-app rounded-lg p-1">
        <button
          class="flex-1 text-center text-2xs py-1 px-4 cursor-pointer rounded-md border-0 transition-colors duration-150
            {activeTab === 'status'
            ? 'bg-surface-input text-fg-bright font-semibold'
            : 'bg-transparent text-fg-secondary hover:text-fg-base'}"
          on:click={() => (activeTab = "status")}>Status</button>
        <button
          class="flex-1 text-center text-2xs py-1 px-4 cursor-pointer rounded-md border-0 transition-colors duration-150
            {activeTab === 'config'
            ? 'bg-surface-input text-fg-bright font-semibold'
            : 'bg-transparent text-fg-secondary hover:text-fg-base'}"
          on:click={() => (activeTab = "config")}>Configuration</button>
      </div>
      <div class="flex items-center gap-2">
        <div class="text-2xs text-fg-muted font-mono">{utcTime}</div>
        <!-- Mini-mode toggle button -->
        <button
          class="flex items-center justify-center w-6 h-6 rounded-md border border-stroke-base text-fg-bright hover:bg-surface-input transition-colors duration-150"
          title="Mini mode"
          on:click={enterMiniMode}
        ><i class="fa-solid fa-compress text-xs"></i></button>
      </div>
    </header>

    <CertBanner {certInfo} />

    <main class="flex-1 overflow-y-auto overflow-x-hidden">
      <div class:hidden={activeTab !== "status"}>
        <StatusTab
          {freqMHz} {mode} {split} {freqTxMHz} {modeTx} {statusMsg} {qsoResult}
          {radioEnabled} {rotatorEnabled}
          {rotConnected} {rotAz} {rotEl} {rotFollow}
          {hfAz} {satAz} {satEl}
          on:follow={(e) => setFollow(e.detail)}
          on:park={park}
        />
      </div>
      <div class:hidden={activeTab !== "config"}><ConfigTab /></div>
    </main>
  {/if}
</div>

<style>
  :global(.hidden) {
    display: none !important;
  }
</style>
