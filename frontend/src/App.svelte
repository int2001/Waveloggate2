<script>
  import { onMount, onDestroy } from 'svelte';
  import { WindowSetSize } from '../wailsjs/runtime/runtime.js';
  import StatusTab from './StatusTab.svelte';
  import ConfigTab from './ConfigTab.svelte';

  const TAB_SIZES = { status: 300, config: 450 };
  const WIDTH = 430;

  let activeTab = 'status';
  let utcTime = '';
  let clockInterval;

  function switchTab(tab) {
    activeTab = tab;
    WindowSetSize(WIDTH, TAB_SIZES[tab]);
  }

  function updateClock() {
    const now = new Date();
    utcTime = now.toUTCString().split(' ').slice(4,5)[0] + ' UTC';
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

<div class="app-shell">
  <header>
    <div class="tab-bar">
      <button class="tab-btn" class:active={activeTab === 'status'}
        on:click={() => switchTab('status')}>Status</button>
      <button class="tab-btn" class:active={activeTab === 'config'}
        on:click={() => switchTab('config')}>Configuration</button>
    </div>
    <div class="clock">{utcTime}</div>
  </header>

  <main>
    <div class:hidden={activeTab !== 'status'}><StatusTab /></div>
    <div class:hidden={activeTab !== 'config'}><ConfigTab /></div>
  </main>
</div>

<style>
  .app-shell {
    display: flex;
    flex-direction: column;
    height: 100vh;
  }

  header {
    background: #1c1c1c;
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 8px;
    height: 32px;
    flex-shrink: 0;
    border-bottom: 1px solid #444;
  }

  .tab-bar {
    display: flex;
    gap: 2px;
  }

  .tab-btn {
    background: transparent;
    border: none;
    border-bottom: 2px solid transparent;
    color: #999;
    padding: 4px 12px;
    font-size: 12px;
    cursor: pointer;
    border-radius: 0;
  }

  .tab-btn:hover {
    background: transparent;
    color: #c6c6c6;
  }

  .tab-btn.active {
    color: #e0e0e0;
    border-bottom-color: #5a9fd4;
  }

  .clock {
    font-size: 11px;
    color: #888;
    font-family: monospace;
  }

  main {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
  }

  :global(.hidden) {
    display: none !important;
  }
</style>
