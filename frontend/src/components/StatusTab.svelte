<script>
  import { onMount, onDestroy } from "svelte";
  import { EventsOn } from "../../wailsjs/runtime/runtime.js";
  import {
    GetConfig,
    GetRotatorStatus,
    RotatorSetFollow,
    RotatorPark,
  } from "../../wailsjs/go/main/App.js";
  import TrxDisplay from "./status/TrxDisplay.svelte";
  import RotatorPanel from "./status/RotatorPanel.svelte";

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

  let offRadio, offQso, offStatus, offRotPos, offRotStatus, offRotBearing, offProfile;

  onMount(async () => {
    offRadio = EventsOn("radio:status", (data) => {
      if (data && data.freqMHz !== undefined) {
        freqMHz = Number(data.freqMHz).toFixed(6);
        mode = data.mode || "";
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

    offProfile = EventsOn("profile:switched", (newRotatorHost) => {
      rotatorHost = newRotatorHost || "";
      // Reset bearing/follow state when switching profiles.
      hfAz = null; satAz = null; satEl = null;
      rotFollow = "off"; rotConnected = false;
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

  async function setFollow(followMode) {
    rotFollow = followMode;
    await RotatorSetFollow(followMode);
  }

  async function park() {
    rotFollow = "off";
    await RotatorPark();
  }
</script>

<div class="py-2.5 px-3 flex flex-col gap-2">
  <TrxDisplay {freqMHz} {mode} />

  {#if statusMsg}
    <div class="alert alert-info font-mono text-2xs">{statusMsg}</div>
  {/if}

  {#if qsoResult}
    {#if qsoResult.success}
      <div class="alert alert-success">
        ✓ QSO logged: <strong>{qsoResult.call}</strong>
        {qsoResult.band} {qsoResult.mode} {qsoResult.rstSent}/{qsoResult.rstRcvd} {qsoResult.timeOn}
      </div>
    {:else}
      <div class="alert alert-danger">
        ✗ QSO NOT logged: {qsoResult.reason || "unknown error"}
      </div>
    {/if}
  {/if}

  {#if rotatorHost}
    <RotatorPanel
      {rotConnected} {rotAz} {rotEl} {rotFollow} {hfAz} {satAz} {satEl}
      on:follow={(e) => setFollow(e.detail)}
      on:park={park}
    />
  {/if}
</div>
