<script>
  import { onMount } from "svelte";
  import {
    GetConfig,
    SaveConfig,
    TestWavelog,
    GetStations,
  } from "../../wailsjs/go/main/App.js";
  import WavelogSection from "./config/WavelogSection.svelte";
  import RadioSection from "./config/RadioSection.svelte";
  import ProfileModal from "./config/ProfileModal.svelte";
  import AdvancedModal from "./config/AdvancedModal.svelte";
  import RotatorModal from "./config/RotatorModal.svelte";

  let cfg = null;
  let stations = [];
  let saveMsg = "";
  let testMsg = "";
  let testSuccess = null;
  let loading = true;

  let showProfileModal = false;
  let showAdvancedModal = false;
  let showRotatorModal = false;

  onMount(async () => {
    cfg = await GetConfig();
    loading = false;
    if (cfg.profiles && cfg.profiles.length > 0) loadStations();
  });

  function activeProfile() {
    return cfg.profiles[cfg.profile] || cfg.profiles[0];
  }

  function setProfileField(key, value) {
    cfg.profiles[cfg.profile][key] = value;
    cfg = cfg; // trigger reactivity
  }

  async function loadStations() {
    const p = activeProfile();
    if (!p || !p.wavelog_url || !p.wavelog_key) return;
    stations = await GetStations(p.wavelog_url, p.wavelog_key);
    cfg = cfg; // force re-evaluation of selected= expressions after options populate
  }

  async function reloadConfig() {
    cfg = await GetConfig();
    stations = [];
    loadStations();
  }

  async function save() {
    cfg = await SaveConfig(cfg);
    saveMsg = "Saved ✓";
    setTimeout(() => (saveMsg = ""), 3000);
  }

  async function test() {
    testMsg = "Testing…";
    testSuccess = null;
    const result = await TestWavelog(activeProfile());
    testSuccess = result.success;
    testMsg = result.success ? "Connection OK ✓" : "Failed: " + result.reason;
    setTimeout(() => { testMsg = ""; testSuccess = null; }, 5000);
  }

  function quit() {
    import("../../wailsjs/runtime/runtime.js").then((r) => r.Quit());
  }

  $: radioType = cfg
    ? cfg.profiles[cfg.profile]?.flrig_ena
      ? "flrig"
      : cfg.profiles[cfg.profile]?.hamlib_ena
        ? "hamlib"
        : "none"
    : "none";

  function setRadioType(type) {
    setProfileField("flrig_ena", type === "flrig");
    setProfileField("hamlib_ena", type === "hamlib");
  }
</script>

{#if loading}
  <div class="p-5 text-fg-muted text-center">Loading…</div>
{:else if cfg}
  <div class="py-2 px-3 flex flex-col gap-1.5">
    <!-- Profile indicator -->
    <div class="flex items-center gap-1.5 bg-surface-card border border-stroke-section px-2 py-0.5 rounded text-2xs">
      <span class="text-fg-muted">Profile:</span>
      <span class="text-accent-value">{cfg.profileNames[cfg.profile] || "Profile " + (cfg.profile + 1)}</span>
    </div>

    <!-- {#key cfg.profile} forces all inputs to be recreated when the active profile changes,
       preventing stale browser input state from leaking across profile switches. -->
    {#key cfg.profile}
      <WavelogSection
        profile={activeProfile()}
        {stations}
        on:fieldchange={(e) => setProfileField(e.detail.key, e.detail.value)}
        on:reloadstations={loadStations}
      />
      <RadioSection
        profile={activeProfile()}
        {radioType}
        on:fieldchange={(e) => setProfileField(e.detail.key, e.detail.value)}
        on:typechange={(e) => setRadioType(e.detail)}
      />
    {/key}

    <!-- Bottom buttons -->
    <div class="mt-1 border-t border-stroke-section pt-2 flex flex-col items-center gap-1">
      <div class="flex gap-1.5 flex-wrap">
        <button on:click={save}>💾 Save</button>
        <button on:click={() => (showProfileModal = true)}>Profiles</button>
        <button on:click={test}>Test</button>
        <button on:click={() => (showAdvancedModal = true)}>⚙ Advanced</button>
        <button on:click={() => (showRotatorModal = true)}>Rot</button>
        <button on:click={quit}>Quit</button>
      </div>

      {#if saveMsg}
        <div class="alert alert-success text-2xs py-0.5 px-2 inline-block">{saveMsg}</div>
      {/if}
      {#if testMsg}
        <div
          class="alert inline-block py-0.5 px-2"
          class:alert-success={testSuccess}
          class:alert-danger={testSuccess === false}
          class:alert-info={testSuccess === null}
        >{testMsg}</div>
      {/if}
    </div>
  </div>
{/if}

{#if showProfileModal && cfg}
  <ProfileModal
    {cfg}
    on:close={() => (showProfileModal = false)}
    on:configchanged={reloadConfig}
  />
{/if}

{#if showAdvancedModal}
  <AdvancedModal on:close={() => (showAdvancedModal = false)} />
{/if}

{#if showRotatorModal && cfg}
  <RotatorModal
    profile={activeProfile()}
    on:fieldchange={(e) => setProfileField(e.detail.key, e.detail.value)}
    on:close={() => (showRotatorModal = false)}
  />
{/if}
