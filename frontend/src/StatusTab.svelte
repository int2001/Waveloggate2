<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';
  import { GetConfig, GetRotatorStatus, RotatorSetFollow, RotatorPark } from '../wailsjs/go/main/App.js';

  let freqMHz = '';
  let mode = '';
  let statusMsg = '';
  let qsoResult = null;  // { success, call, band, mode, rstSent, rstRcvd, timeOn, reason }

  // Rotator state
  let rotatorHost = '';
  let rotConnected = false;
  let rotAz = 0;
  let rotEl = 0;
  let rotFollow = 'off';
  let hfAz = null;
  let satAz = null;
  let satEl = null;

  let offRadio, offQso, offStatus, offRotPos, offRotStatus, offRotBearing, offProfile;

  onMount(async () => {
    offRadio = EventsOn('radio:status', (data) => {
      if (data && data.freqMHz !== undefined) {
        freqMHz = Number(data.freqMHz).toFixed(6);
        mode = data.mode || '';
      }
    });

    offQso = EventsOn('qso:result', (data) => {
      qsoResult = data;
      setTimeout(() => { qsoResult = null; }, 30000);
    });

    offStatus = EventsOn('status:message', (msg) => {
      statusMsg = msg;
    });

    offRotPos = EventsOn('rotator:position', (data) => {
      if (data) { rotAz = data.az; rotEl = data.el; }
    });

    offRotStatus = EventsOn('rotator:status', (connected) => {
      rotConnected = connected;
    });

    offRotBearing = EventsOn('rotator:bearing', (data) => {
      if (!data) return;
      if (data.type === 'hf') { hfAz = data.az; }
      if (data.type === 'sat') { satAz = data.az; satEl = data.el; }
    });

    offProfile = EventsOn('profile:switched', (newRotatorHost) => {
      rotatorHost = newRotatorHost || '';
      // Reset bearing/follow state when switching profiles.
      hfAz = null; satAz = null; satEl = null;
      rotFollow = 'off';
      rotConnected = false;
    });

    // Load initial state.
    const cfg = await GetConfig();
    rotatorHost = cfg.profiles?.[cfg.profile]?.rotator_host || '';
    if (rotatorHost) {
      const s = await GetRotatorStatus();
      rotConnected = s.connected;
      rotAz = s.az;
      rotEl = s.el;
      rotFollow = s.followMode || 'off';
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
    rotFollow = 'off';
    await RotatorPark();
  }
</script>

<div class="status-tab">
  <div class="trx-display">
    <span class="label">TRX:</span>
    {#if freqMHz}
      <span class="freq">{freqMHz} MHz</span>
      <span class="sep">/</span>
      <span class="mode">{mode}</span>
    {:else}
      <span class="no-radio">No radio data</span>
    {/if}
  </div>

  {#if statusMsg}
    <div class="alert alert-info status-msg">{statusMsg}</div>
  {/if}

  {#if qsoResult}
    {#if qsoResult.success}
      <div class="alert alert-success">
        ✓ QSO logged: <strong>{qsoResult.call}</strong>
        {qsoResult.band} {qsoResult.mode}
        {qsoResult.rstSent}/{qsoResult.rstRcvd}
        {qsoResult.timeOn}
      </div>
    {:else}
      <div class="alert alert-danger">
        ✗ QSO NOT logged: {qsoResult.reason || 'unknown error'}
      </div>
    {/if}
  {/if}

  {#if rotatorHost}
  <div class="rot-panel">
    <div class="rot-header">
      <span class="label">ROTATOR</span>
      <span class="rot-dot" class:connected={rotConnected}></span>
      <span class="rot-status">{rotConnected ? 'connected' : 'disconnected'}</span>
    </div>

    <div class="rot-pos">
      <span class="rot-field">Az: <span class="rot-val">{rotAz.toFixed(1)}°</span></span>
      <span class="rot-field">El: <span class="rot-val">{rotEl.toFixed(1)}°</span></span>
    </div>

    <div class="rot-follow">
      <label class="rot-radio">
        <input type="radio" name="follow" value="off" checked={rotFollow === 'off'}
          on:change={() => setFollow('off')} /> Off
      </label>
      <label class="rot-radio">
        <input type="radio" name="follow" value="hf" checked={rotFollow === 'hf'}
          on:change={() => setFollow('hf')} /> HF
        {#if hfAz !== null}<span class="rot-bearing">Az: {hfAz.toFixed(0)}°</span>{/if}
      </label>
      <label class="rot-radio">
        <input type="radio" name="follow" value="sat" checked={rotFollow === 'sat'}
          on:change={() => setFollow('sat')} /> SAT
        {#if satAz !== null}<span class="rot-bearing">Az: {satAz.toFixed(0)}°  El: {satEl.toFixed(0)}°</span>{/if}
      </label>
    </div>

    <div class="rot-actions">
      <button class="rot-park-btn" on:click={park}>Park</button>
    </div>
  </div>
  {/if}
</div>

<style>
  .status-tab {
    padding: 10px 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .trx-display {
    background: #262626;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 8px 12px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-family: monospace;
  }

  .label {
    color: #888;
    font-size: 11px;
  }

  .freq {
    color: #5af;
    font-size: 16px;
    font-weight: bold;
  }

  .sep {
    color: #666;
  }

  .mode {
    color: #fa5;
    font-size: 14px;
  }

  .no-radio {
    color: #666;
    font-size: 12px;
  }

  .status-msg {
    font-family: monospace;
    font-size: 11px;
  }

  .rot-panel {
    background: #262626;
    border: 1px solid #444;
    border-radius: 4px;
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-family: monospace;
  }

  .rot-header {
    display: flex;
    align-items: center;
    gap: 6px;
  }

  .rot-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: #666;
    flex-shrink: 0;
  }

  .rot-dot.connected {
    background: #4c4;
  }

  .rot-status {
    color: #888;
    font-size: 11px;
  }

  .rot-pos {
    display: flex;
    gap: 16px;
  }

  .rot-field {
    color: #888;
    font-size: 12px;
  }

  .rot-val {
    color: #5af;
    font-weight: bold;
  }

  .rot-follow {
    display: flex;
    gap: 12px;
    align-items: center;
    flex-wrap: wrap;
  }

  .rot-radio {
    display: flex;
    align-items: center;
    gap: 4px;
    color: #aaa;
    font-size: 12px;
    cursor: pointer;
  }

  .rot-bearing {
    color: #fa5;
    font-size: 11px;
    margin-left: 4px;
  }

  .rot-actions {
    display: flex;
  }

  .rot-park-btn {
    font-size: 11px;
    padding: 2px 10px;
  }
</style>
