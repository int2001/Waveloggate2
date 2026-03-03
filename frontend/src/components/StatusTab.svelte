<script>
  import { onMount, onDestroy } from "svelte";
  import { EventsOn } from "../../wailsjs/runtime/runtime.js";
  import {
    GetConfig,
    GetRotatorStatus,
    RotatorSetFollow,
    RotatorPark,
  } from "../../wailsjs/go/main/App.js";

  let freqMHz = "";
  let mode = "";
  let statusMsg = "";
  let qsoResult = null; // { success, call, band, mode, rstSent, rstRcvd, timeOn, reason }

  // Rotator state
  let rotatorHost = "";
  let rotConnected = false;
  let rotAz = 0;
  let rotEl = 0;
  let rotFollow = "off";
  let hfAz = null;
  let satAz = null;
  let satEl = null;

  let offRadio,
    offQso,
    offStatus,
    offRotPos,
    offRotStatus,
    offRotBearing,
    offProfile;

  onMount(async () => {
    offRadio = EventsOn("radio:status", (data) => {
      if (data && data.freqMHz !== undefined) {
        freqMHz = Number(data.freqMHz).toFixed(6);
        mode = data.mode || "";
      }
    });

    offQso = EventsOn("qso:result", (data) => {
      qsoResult = data;
      setTimeout(() => {
        qsoResult = null;
      }, 30000);
    });

    offStatus = EventsOn("status:message", (msg) => {
      statusMsg = msg;
    });

    offRotPos = EventsOn("rotator:position", (data) => {
      if (data) {
        rotAz = data.az;
        rotEl = data.el;
      }
    });

    offRotStatus = EventsOn("rotator:status", (connected) => {
      rotConnected = connected;
    });

    offRotBearing = EventsOn("rotator:bearing", (data) => {
      if (!data) return;
      if (data.type === "hf") {
        hfAz = data.az;
      }
      if (data.type === "sat") {
        satAz = data.az;
        satEl = data.el;
      }
    });

    offProfile = EventsOn("profile:switched", (newRotatorHost) => {
      rotatorHost = newRotatorHost || "";
      // Reset bearing/follow state when switching profiles.
      hfAz = null;
      satAz = null;
      satEl = null;
      rotFollow = "off";
      rotConnected = false;
    });

    // Load initial state.
    const cfg = await GetConfig();
    rotatorHost = cfg.profiles?.[cfg.profile]?.rotator_host || "";
    if (rotatorHost) {
      const s = await GetRotatorStatus();
      rotConnected = s.connected;
      rotAz = s.az;
      rotEl = s.el;
      rotFollow = s.followMode || "off";
    }
  });

  onDestroy(() => {
    if (offRadio) offRadio();
    if (offQso) offQso();
    if (offStatus) offStatus();
    if (offRotPos) offRotPos();
    if (offRotStatus) offRotStatus();
    if (offRotBearing) offRotBearing();
    if (offProfile) offProfile();
  });

  async function setFollow(mode) {
    rotFollow = mode;
    await RotatorSetFollow(mode);
  }

  async function park() {
    rotFollow = "off";
    await RotatorPark();
  }
</script>

<div class="py-2.5 px-3 flex flex-col gap-2">
  <!-- TRX display -->
  <div
    class="bg-surface-card border border-stroke-subtle rounded px-3 py-2 flex items-center gap-2 font-mono"
  >
    <span class="text-fg-muted text-2xs">TRX:</span>
    {#if freqMHz}
      <span class="text-accent-value text-base font-bold">{freqMHz} MHz</span>
      <span class="text-fg-dim">/</span>
      <span class="text-accent-orange text-sm">{mode}</span>
    {:else}
      <span class="text-fg-dim text-xs">No radio data</span>
    {/if}
  </div>

  {#if statusMsg}
    <div class="alert alert-info font-mono text-2xs">{statusMsg}</div>
  {/if}

  {#if qsoResult}
    {#if qsoResult.success}
      <div class="alert alert-success">
        ✓ QSO logged: <strong>{qsoResult.call}</strong>
        {qsoResult.band}
        {qsoResult.mode}
        {qsoResult.rstSent}/{qsoResult.rstRcvd}
        {qsoResult.timeOn}
      </div>
    {:else}
      <div class="alert alert-danger">
        ✗ QSO NOT logged: {qsoResult.reason || "unknown error"}
      </div>
    {/if}
  {/if}

  {#if rotatorHost}
    <!-- Rotator panel -->
    <div
      class="bg-surface-card border border-stroke-subtle rounded px-3 py-2 flex flex-col gap-1.5 font-mono"
    >
      <div class="flex items-center gap-1.5">
        <span class="text-fg-muted text-2xs">ROTATOR</span>
        <span
          class="w-2 h-2 rounded-full flex-shrink-0 {rotConnected
            ? 'bg-accent-green'
            : 'bg-fg-dim'}"
        ></span>
        <span class="text-fg-muted text-2xs"
          >{rotConnected ? "connected" : "disconnected"}</span
        >
      </div>

      <div class="flex gap-4">
        <span class="text-fg-muted text-xs"
          >Az: <span class="text-accent-value font-bold"
            >{rotAz.toFixed(1)}°</span
          ></span
        >
        <span class="text-fg-muted text-xs"
          >El: <span class="text-accent-value font-bold"
            >{rotEl.toFixed(1)}°</span
          ></span
        >
      </div>

      <div class="flex gap-3 items-center flex-wrap">
        <label
          class="flex items-center gap-1 text-fg-label text-xs cursor-pointer"
        >
          <input
            type="radio"
            name="follow"
            value="off"
            checked={rotFollow === "off"}
            on:change={() => setFollow("off")}
          /> Off
        </label>
        <label
          class="flex items-center gap-1 text-fg-label text-xs cursor-pointer"
        >
          <input
            type="radio"
            name="follow"
            value="hf"
            checked={rotFollow === "hf"}
            on:change={() => setFollow("hf")}
          />
          HF
          {#if hfAz !== null}<span class="text-accent-orange text-2xs ml-1"
              >Az: {hfAz.toFixed(0)}°</span
            >{/if}
        </label>
        <label
          class="flex items-center gap-1 text-fg-label text-xs cursor-pointer"
        >
          <input
            type="radio"
            name="follow"
            value="sat"
            checked={rotFollow === "sat"}
            on:change={() => setFollow("sat")}
          />
          SAT
          {#if satAz !== null}<span class="text-accent-orange text-2xs ml-1"
              >Az: {satAz.toFixed(0)}° El: {satEl.toFixed(0)}°</span
            >{/if}
        </label>
      </div>

      <div class="flex">
        <button class="text-2xs py-0.5 px-2.5" on:click={park}>Park</button>
      </div>
    </div>
  {/if}
</div>
