<script>
  export let freqMHz = "";
  export let mode = "";
</script>

<div class="bg-surface-card border border-stroke-subtle rounded-lg px-4 py-3 font-mono">
  <div class="flex items-start justify-between">
    <div>
      <div class="text-fg-muted text-2xs uppercase tracking-widest mb-2">Transceiver</div>
      {#if freqMHz}
        <div class="text-accent-value text-2xl font-bold tracking-tight leading-none">{freqMHz}</div>
        <div class="text-fg-muted text-xs mt-1.5">MHz</div>
      {:else}
        <div class="text-fg-dim text-sm italic mt-2">No radio data</div>
      {/if}
    </div>
    {#if mode}
      <div class="bg-surface-app border border-stroke-section text-accent-orange text-l font-semibold px-2.5 py-1 rounded-md mt-5">
        {mode}
      </div>
    {/if}
  </div>
</div>
