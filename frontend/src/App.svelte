<script>
  import { onMount, onDestroy } from "svelte";
  import { WindowSetSize } from "../wailsjs/runtime/runtime.js";
  import StatusTab from "./components/StatusTab.svelte";
  import ConfigTab from "./components/ConfigTab.svelte";

  const TAB_SIZES = { status: 380, config: 450 };
  const WIDTH = 430;

  let activeTab = "status";
  let utcTime = "";
  let clockInterval;

  function switchTab(tab) {
    activeTab = tab;
    WindowSetSize(WIDTH, TAB_SIZES[tab]);
  }

  function updateClock() {
    const now = new Date();
    utcTime = now.toUTCString().split(" ").slice(4, 5)[0] + " UTC";
  }

  onMount(() => {
    WindowSetSize(WIDTH, TAB_SIZES[activeTab]);
    updateClock();
    clockInterval = setInterval(updateClock, 1000);
  });

  onDestroy(() => {
    clearInterval(clockInterval);
  });
</script>

<div class="flex flex-col h-screen">
  <header
    class="bg-surface-header flex items-center justify-between px-2 h-8 flex-shrink-0 border-b border-stroke-subtle"
  >
    <div class="flex gap-0.5">
      <button
        class="bg-transparent border-0 border-b-2 text-2xs py-1 px-3 cursor-pointer rounded-none transition-colors duration-100
          {activeTab === 'status'
          ? 'text-fg-bright border-b-stroke-accent font-medium'
          : 'text-fg-secondary border-b-transparent hover:text-fg-base'}"
        on:click={() => switchTab("status")}>Status</button
      >
      <button
        class="bg-transparent border-0 border-b-2 text-2xs py-1 px-3 cursor-pointer rounded-none transition-colors duration-100
          {activeTab === 'config'
          ? 'text-fg-bright border-b-stroke-accent font-medium'
          : 'text-fg-secondary border-b-transparent hover:text-fg-base'}"
        on:click={() => switchTab("config")}>Configuration</button
      >
    </div>
    <div class="text-2xs text-fg-muted font-mono">{utcTime}</div>
  </header>

  <main class="flex-1 overflow-y-auto overflow-x-hidden">
    <div class:hidden={activeTab !== "status"}><StatusTab /></div>
    <div class:hidden={activeTab !== "config"}><ConfigTab /></div>
  </main>
</div>

<style>
  :global(.hidden) {
    display: none !important;
  }
</style>
