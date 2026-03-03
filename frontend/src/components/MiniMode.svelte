<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let utcTime = "";
  export let freqMHz = "";
  export let mode = "";
  export let qsoResult = null;
  export let rotatorEnabled = false;
  export let minimapEnabled = false;
  export let rotConnected = false;
  export let rotAz = 0;
  export let rotEl = 0;
  export let rotFollow = "off";
  export let hfAz = null;
  export let satAz = null;
  export let satEl = null;

  // Polar map geometry (50px radius for 110×110 SVG)
  $: azRad = (rotAz - 90) * Math.PI / 180;
  $: elR   = 50 * (1 - rotEl / 90);
</script>

<!-- Mini header: expand icon (left) + UTC clock (right) -->
<header class="bg-surface-header flex items-center justify-between px-3 h-10 flex-shrink-0 border-b border-stroke-subtle">
  <button
    class="flex items-center justify-center w-6 h-6 rounded-md border border-stroke-base text-fg-bright hover:bg-surface-input transition-colors duration-150"
    title="Back to full view"
    on:click={() => dispatch("expand")}
  ><i class="fa-solid fa-expand text-xs"></i></button>
  <div class="text-2xs text-fg-muted font-mono">{utcTime}</div>
</header>

<!-- Compact body -->
<div class="px-3 pt-2 pb-2.5 flex flex-col gap-1.5">

  <!-- Row 1: Frequency + Mode badge -->
  <div class="flex items-center gap-2 font-mono">
    {#if freqMHz}
      <span class="text-accent-value text-xl font-bold leading-none">{freqMHz}</span>
      <span class="text-fg-muted text-2xs self-end mb-0.5">MHz</span>
    {:else}
      <span class="text-fg-dim text-xs italic">No radio data</span>
    {/if}
    {#if mode}
      <span class="ml-auto bg-surface-app border border-stroke-section text-accent-orange text-xs font-semibold px-2 py-0.5 rounded-md">
        {mode}
      </span>
    {/if}
  </div>

  <!-- Row 2: QSO result (single compact line) -->
  <div class="text-2xs leading-tight min-h-[16px]">
    {#if qsoResult}
      {#if qsoResult.success}
        <span class="text-accent-green">✓</span>
        <span class="text-fg-base font-semibold">{qsoResult.call}</span>
        <span class="text-fg-muted">&nbsp;{qsoResult.band} {qsoResult.mode} {qsoResult.rstSent}/{qsoResult.rstRcvd} {qsoResult.timeOn}</span>
      {:else}
        <span class="text-red-400">✗</span>
        <span class="text-fg-muted">&nbsp;{qsoResult.reason || "log error"}</span>
      {/if}
    {:else}
      <span class="text-fg-dim italic">–</span>
    {/if}
  </div>

  <!-- Rotator section (only when enabled) -->
  {#if rotatorEnabled}
    <!-- Az / El + connection dot -->
    <div class="flex items-center gap-3 font-mono border-t border-stroke-section pt-1.5">
      <div class="flex items-center gap-1">
        <span class="text-fg-muted text-2xs uppercase tracking-wider">Az</span>
        <span class="text-accent-value text-sm font-bold">{rotAz.toFixed(1)}°</span>
      </div>
      <div class="flex items-center gap-1">
        <span class="text-fg-muted text-2xs uppercase tracking-wider">El</span>
        <span class="text-accent-value text-sm font-bold">{rotEl.toFixed(1)}°</span>
      </div>
      <div class="ml-auto flex items-center gap-1">
        <span class="w-1.5 h-1.5 rounded-full flex-shrink-0 {rotConnected ? 'bg-accent-green' : 'bg-fg-dim'}"></span>
      </div>
    </div>

    <!-- Follow mode segmented control -->
    <div class="flex rounded-md overflow-hidden border border-stroke-section">
      <button
        class="flex-1 py-1 text-2xs font-medium border-0 rounded-none transition-colors duration-100
          {rotFollow === 'off' ? 'bg-surface-input text-fg-bright' : 'bg-surface-app text-fg-secondary hover:bg-surface-section hover:text-fg-base'}"
        on:click={() => dispatch("follow", "off")}
      >Off</button>
      <button
        class="flex-1 py-1 text-2xs font-medium border-0 border-l border-stroke-section rounded-none transition-colors duration-100
          {rotFollow === 'hf' ? 'bg-surface-input text-fg-bright' : 'bg-surface-app text-fg-secondary hover:bg-surface-section hover:text-fg-base'}"
        on:click={() => dispatch("follow", "hf")}
      >HF{#if hfAz !== null}&nbsp;<span class="text-accent-orange">→{Number(hfAz).toFixed(0)}°</span>{/if}</button>
      <button
        class="flex-1 py-1 text-2xs font-medium border-0 border-l border-stroke-section rounded-none transition-colors duration-100
          {rotFollow === 'sat' ? 'bg-surface-input text-fg-bright' : 'bg-surface-app text-fg-secondary hover:bg-surface-section hover:text-fg-base'}"
        on:click={() => dispatch("follow", "sat")}
      >SAT{#if satAz !== null}&nbsp;<span class="text-accent-orange">↗{Number(satAz).toFixed(0)}°</span>{/if}</button>
    </div>

    <!-- Compact polar map (110×110, only when minimapEnabled) -->
    {#if minimapEnabled}
      <div class="flex justify-center pt-0.5">
        <svg viewBox="0 0 110 110" width="110" height="110" xmlns="http://www.w3.org/2000/svg">
          <!-- Background -->
          <circle cx="55" cy="55" r="50" fill="#1e1e1e" stroke="#404040" stroke-width="1"/>
          <!-- Elevation rings -->
          <circle cx="55" cy="55" r="17" fill="none" stroke="#383838" stroke-width="0.75" stroke-dasharray="2,3"/>
          <circle cx="55" cy="55" r="33" fill="none" stroke="#383838" stroke-width="0.75" stroke-dasharray="2,3"/>
          <!-- Cardinal ticks -->
          {#each [0, 90, 180, 270] as deg}
            {@const rad = (deg - 90) * Math.PI / 180}
            <line
              x1={55 + 46 * Math.cos(rad)} y1={55 + 46 * Math.sin(rad)}
              x2={55 + 50 * Math.cos(rad)} y2={55 + 50 * Math.sin(rad)}
              stroke="#555555" stroke-width="1"
            />
          {/each}
          <!-- Cardinal labels -->
          <text x="55"  y="3"   text-anchor="middle" dominant-baseline="hanging" fill="#777" font-size="7" font-family="monospace">N</text>
          <text x="107" y="55"  text-anchor="end"    dominant-baseline="middle"  fill="#777" font-size="7" font-family="monospace">E</text>
          <text x="55"  y="107" text-anchor="middle" dominant-baseline="auto"    fill="#777" font-size="7" font-family="monospace">S</text>
          <text x="3"   y="55"  text-anchor="start"  dominant-baseline="middle"  fill="#777" font-size="7" font-family="monospace">W</text>
          <!-- HF bearing -->
          {#if rotFollow === "hf" && hfAz !== null}
            {@const r2 = (Number(hfAz) - 90) * Math.PI / 180}
            <line
              x1="55" y1="55"
              x2={55 + 48 * Math.cos(r2)} y2={55 + 48 * Math.sin(r2)}
              stroke="#ffaa55" stroke-width="1.5" stroke-dasharray="4,3" stroke-linecap="round" opacity="0.65"
            />
          {/if}
          <!-- SAT position -->
          {#if rotFollow === "sat" && satAz !== null}
            {@const r2 = (Number(satAz) - 90) * Math.PI / 180}
            {@const rEl = 50 * (1 - Number(satEl || 0) / 90)}
            <circle cx={55 + rEl * Math.cos(r2)} cy={55 + rEl * Math.sin(r2)} r="3" fill="#ffaa55" opacity="0.8"/>
          {/if}
          <!-- Rotator needle + position dot -->
          <line
            x1="55" y1="55"
            x2={55 + 48 * Math.cos(azRad)} y2={55 + 48 * Math.sin(azRad)}
            stroke="#55aaff" stroke-width="1" opacity="0.25" stroke-linecap="round"
          />
          <circle cx={55 + elR * Math.cos(azRad)} cy={55 + elR * Math.sin(azRad)} r="4" fill="#55aaff"/>
          <!-- Zenith dot -->
          <circle cx="55" cy="55" r="2" fill="#555555"/>
        </svg>
      </div>
    {/if}
  {/if}

</div>
