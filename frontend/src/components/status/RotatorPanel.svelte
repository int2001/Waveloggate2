<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let rotConnected = false;
  export let rotAz = 0;
  export let rotEl = 0;
  export let rotFollow = "off";
  export let hfAz = null;
  export let satAz = null;
  export let satEl = null;
</script>

<div class="bg-surface-card border border-stroke-subtle rounded-lg px-4 py-3 flex flex-col gap-3 font-mono">

  <!-- Header -->
  <div class="flex items-center justify-between">
    <span class="text-fg-muted text-2xs uppercase tracking-widest font-semibold">Rotator</span>
    <div class="flex items-center gap-1.5">
      <span class="w-1.5 h-1.5 rounded-full flex-shrink-0 {rotConnected ? 'bg-accent-green' : 'bg-fg-dim'}"></span>
      <span class="text-fg-dim text-2xs">{rotConnected ? "connected" : "disconnected"}</span>
    </div>
  </div>

  <!-- Az / El instrument tiles -->
  <div class="flex gap-2">
    <div class="flex-1 bg-surface-app border border-stroke-section rounded-md px-3 py-2.5 flex flex-col gap-0.5">
      <span class="text-fg-muted text-2xs uppercase tracking-wider">Azimuth</span>
      <span class="text-accent-value text-xl font-bold leading-tight">{rotAz.toFixed(1)}°</span>
    </div>
    <div class="flex-1 bg-surface-app border border-stroke-section rounded-md px-3 py-2.5 flex flex-col gap-0.5">
      <span class="text-fg-muted text-2xs uppercase tracking-wider">Elevation</span>
      <span class="text-accent-value text-xl font-bold leading-tight">{rotEl.toFixed(1)}°</span>
    </div>
  </div>

  <!-- Follow mode: segmented control -->
  <div class="flex flex-col gap-1">
    <span class="text-fg-muted text-2xs uppercase tracking-wider">Follow Mode</span>
    <div class="flex rounded-md overflow-hidden border border-stroke-section">
      <button
        class="flex-1 py-1.5 text-xs font-medium border-0 rounded-none transition-colors duration-100
          {rotFollow === 'off'
          ? 'bg-surface-input text-fg-bright hover:bg-surface-input'
          : 'bg-surface-app text-fg-secondary hover:bg-surface-section hover:text-fg-base'}"
        on:click={() => dispatch("follow", "off")}
      >Off</button>
      <button
        class="flex-1 py-1.5 text-xs font-medium border-0 border-l border-stroke-section rounded-none transition-colors duration-100
          {rotFollow === 'hf'
          ? 'bg-surface-input text-fg-bright hover:bg-surface-input'
          : 'bg-surface-app text-fg-secondary hover:bg-surface-section hover:text-fg-base'}"
        on:click={() => dispatch("follow", "hf")}
      >HF{#if hfAz !== null} <span class="text-accent-orange text-2xs">→{hfAz.toFixed(0)}°</span>{/if}</button>
      <button
        class="flex-1 py-1.5 text-xs font-medium border-0 border-l border-stroke-section rounded-none transition-colors duration-100
          {rotFollow === 'sat'
          ? 'bg-surface-input text-fg-bright hover:bg-surface-input'
          : 'bg-surface-app text-fg-secondary hover:bg-surface-section hover:text-fg-base'}"
        on:click={() => dispatch("follow", "sat")}
      >SAT{#if satAz !== null} <span class="text-accent-orange text-2xs">↗{satAz.toFixed(0)}°</span>{/if}</button>
    </div>
  </div>

  <!-- Park -->
  <div class="flex justify-end">
    <button
      class="text-xs py-1.5 px-4 text-fg-secondary hover:text-fg-base"
      on:click={() => dispatch("park")}
    >Park ⟳</button>
  </div>

</div>
