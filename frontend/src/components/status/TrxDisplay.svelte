<script>
  export let freqMHz = "";
  export let mode = "";
</script>

<div class="bg-surface-card border border-stroke-subtle rounded px-3 py-2 flex items-center gap-2.5 font-mono">
  <span class="text-fg-muted text-2xs pr-2.5 border-r border-stroke-section">TRX</span>
  {#if freqMHz}
    <span class="text-accent-value text-base font-bold tracking-wide">{freqMHz} MHz</span>
    <span class="text-fg-dim text-xs">·</span>
    <span class="text-accent-orange text-sm font-semibold">{mode}</span>
  {:else}
    <span class="text-fg-dim text-xs italic">No radio data</span>
  {/if}
</div>
