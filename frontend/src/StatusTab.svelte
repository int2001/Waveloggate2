<script>
  import { onMount, onDestroy } from 'svelte';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';

  let freqMHz = '';
  let mode = '';
  let statusMsg = '';
  let qsoResult = null;  // { success, call, band, mode, rstSent, rstRcvd, timeOn, reason }

  let offRadio, offQso, offStatus;

  onMount(() => {
    offRadio = EventsOn('radio:status', (data) => {
      if (data && data.freqMHz !== undefined) {
        freqMHz = Number(data.freqMHz).toFixed(6);
        mode = data.mode || '';
      }
    });

    offQso = EventsOn('qso:result', (data) => {
      qsoResult = data;
      // Auto-clear after 30s
      setTimeout(() => { qsoResult = null; }, 30000);
    });

    offStatus = EventsOn('status:message', (msg) => {
      statusMsg = msg;
    });
  });

  onDestroy(() => {
    if (offRadio) offRadio();
    if (offQso) offQso();
    if (offStatus) offStatus();
  });
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
</style>
