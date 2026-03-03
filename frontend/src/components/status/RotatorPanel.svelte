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

<div class="bg-surface-card border border-stroke-subtle rounded px-3 py-2 flex flex-col gap-1.5 font-mono">
  <div class="flex items-center gap-1.5">
    <span class="text-fg-muted text-2xs">ROTATOR</span>
    <span class="w-2 h-2 rounded-full flex-shrink-0 {rotConnected ? 'bg-accent-green' : 'bg-fg-dim'}"></span>
    <span class="text-fg-muted text-2xs">{rotConnected ? "connected" : "disconnected"}</span>
  </div>

  <div class="flex gap-5">
    <div class="flex flex-col">
      <span class="text-fg-muted text-2xs">Az</span>
      <span class="text-accent-value font-bold text-sm font-mono">{rotAz.toFixed(1)}°</span>
    </div>
    <div class="flex flex-col">
      <span class="text-fg-muted text-2xs">El</span>
      <span class="text-accent-value font-bold text-sm font-mono">{rotEl.toFixed(1)}°</span>
    </div>
  </div>

  <div class="flex gap-3 items-center flex-wrap">
    <label class="flex items-center gap-1 text-fg-label text-xs cursor-pointer">
      <input type="radio" name="follow" value="off"
        checked={rotFollow === "off"} on:change={() => dispatch("follow", "off")} /> Off
    </label>
    <label class="flex items-center gap-1 text-fg-label text-xs cursor-pointer">
      <input type="radio" name="follow" value="hf"
        checked={rotFollow === "hf"} on:change={() => dispatch("follow", "hf")} />
      HF
      {#if hfAz !== null}<span class="text-accent-orange text-2xs ml-1">→ {hfAz.toFixed(0)}°</span>{/if}
    </label>
    <label class="flex items-center gap-1 text-fg-label text-xs cursor-pointer">
      <input type="radio" name="follow" value="sat"
        checked={rotFollow === "sat"} on:change={() => dispatch("follow", "sat")} />
      SAT
      {#if satAz !== null}<span class="text-accent-orange text-2xs ml-1">↗ {satAz.toFixed(0)}° / {satEl.toFixed(0)}°</span>{/if}
    </label>
  </div>

  <div class="flex">
    <button class="text-2xs py-0.5 px-2.5 text-fg-secondary hover:text-fg-base"
      on:click={() => dispatch("park")}>Park</button>
  </div>
</div>
