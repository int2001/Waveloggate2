<script>
  export let freqMHz = "";
  export let mode = "";
  export let split = false;
  export let freqTxMHz = "";
  export let modeTx = "";
</script>

<div class="bg-surface-card border border-stroke-subtle rounded-lg px-4 py-3 font-mono">
  <div class="text-fg-muted text-2xs uppercase tracking-widest mb-2">Transceiver</div>
  {#if freqMHz}
    {#if split}
      <div class="flex items-start justify-between gap-4">
        <div class="flex items-start gap-4">
          <div>
            <div class="text-fg-muted text-2xs uppercase tracking-wider mb-1">RX</div>
            <div class="text-accent-value text-2xl font-bold tracking-tight leading-none">{freqMHz}</div>
            <div class="text-fg-muted text-xs mt-1.5">MHz</div>
          </div>
          {#if freqTxMHz}
            <div class="border-l border-stroke-section pl-4">
              <div class="text-fg-muted text-2xs uppercase tracking-wider mb-1">TX</div>
              <div class="text-accent-value text-2xl font-bold tracking-tight leading-none">{freqTxMHz}</div>
              <div class="text-fg-muted text-xs mt-1.5">MHz</div>
            </div>
          {/if}
        </div>
        <div class="flex flex-col items-end gap-1.5">
          <span class="bg-accent-orange/10 border border-accent-orange/40 text-accent-orange text-2xs font-bold px-2 py-0.5 rounded-md tracking-wider">SPLIT</span>
          {#if mode}
            <div class="bg-surface-app border border-stroke-section text-accent-orange text-l font-semibold px-2.5 py-1 rounded-md">{mode}</div>
          {/if}
          {#if modeTx && modeTx !== mode}
            <div class="bg-surface-app border border-stroke-section text-accent-orange text-l font-semibold px-2.5 py-1 rounded-md">{modeTx}</div>
          {/if}
        </div>
      </div>
    {:else}
      <div class="flex items-start justify-between">
        <div>
          <div class="text-accent-value text-2xl font-bold tracking-tight leading-none">{freqMHz}</div>
          <div class="text-fg-muted text-xs mt-1.5">MHz</div>
        </div>
        {#if mode}
          <div class="flex flex-col items-end gap-1.5 mt-5">
            <div class="bg-surface-app border border-stroke-section text-accent-orange text-l font-semibold px-2.5 py-1 rounded-md">{mode}</div>
          </div>
        {/if}
      </div>
    {/if}
  {:else}
    <div class="text-fg-dim text-sm italic mt-2">No radio data</div>
  {/if}
</div>
