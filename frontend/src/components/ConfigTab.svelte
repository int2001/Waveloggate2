<script>
  import { onMount } from "svelte";
  import {
    GetConfig,
    SaveConfig,
    TestWavelog,
    GetStations,
    CreateProfile,
    DeleteProfile,
    RenameProfile,
    SwitchProfile,
    SaveAdvanced,
    GetUDPStatus,
  } from "../../wailsjs/go/main/App.js";

  let cfg = null;
  let stations = [];
  let saveMsg = "";
  let testMsg = "";
  let testSuccess = null;
  let loading = true;

  // Modals
  let showProfileModal = false;
  let showAdvancedModal = false;

  // Profile modal state
  let newProfileName = "";
  let renameIndex = -1;
  let renameName = "";

  // Advanced modal state
  let advUdpEnabled = true;
  let advUdpPort = 2333;
  let advStatus = "";

  // Rotator modal
  let showRotatorModal = false;

  onMount(async () => {
    cfg = await GetConfig();
    loading = false;
    if (cfg.profiles && cfg.profiles.length > 0) {
      loadStations();
    }
  });

  function activeProfile() {
    if (!cfg) return null;
    return cfg.profiles[cfg.profile] || cfg.profiles[0];
  }

  function setProfileField(key, value) {
    cfg.profiles[cfg.profile][key] = value;
    cfg = cfg; // reactivity
  }

  async function loadStations() {
    const p = activeProfile();
    if (!p || !p.wavelog_url || !p.wavelog_key) return;
    stations = await GetStations(p.wavelog_url, p.wavelog_key);
    cfg = cfg; // force re-evaluation of selected= expressions after options populate
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
    setTimeout(() => {
      testMsg = "";
      testSuccess = null;
    }, 5000);
  }

  // Profile modal
  async function openProfileModal() {
    showProfileModal = true;
    newProfileName = "";
    renameIndex = -1;
    renameName = "";
  }

  async function doCreateProfile() {
    if (!newProfileName.trim()) return;
    await CreateProfile(newProfileName.trim());
    cfg = await GetConfig();
    newProfileName = "";
  }

  async function doDeleteProfile(i) {
    const err = await DeleteProfile(i);
    if (err) {
      alert(err);
      return;
    }
    cfg = await GetConfig();
  }

  async function doRenameProfile() {
    if (renameIndex < 0 || !renameName.trim()) return;
    await RenameProfile(renameIndex, renameName.trim());
    cfg = await GetConfig();
    renameIndex = -1;
    renameName = "";
  }

  async function doSwitchProfile(i) {
    await SwitchProfile(i);
    cfg = await GetConfig();
    stations = [];
    loadStations();
  }

  // Advanced modal
  async function openAdvancedModal() {
    const status = await GetUDPStatus();
    advUdpEnabled = status.enabled;
    advUdpPort = status.port;
    advStatus = "";
    showAdvancedModal = true;
  }

  async function doSaveAdvanced() {
    const err = await SaveAdvanced(advUdpEnabled, advUdpPort);
    if (err) {
      advStatus = "Error: " + err;
    } else {
      advStatus = "Saved ✓";
      setTimeout(() => {
        advStatus = "";
        showAdvancedModal = false;
      }, 1500);
    }
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
    <div
      class="flex items-center gap-1.5 bg-surface-card px-2 py-0.5 rounded text-2xs"
    >
      <span class="text-fg-muted">Profile:</span>
      <span class="text-accent-value"
        >{cfg.profileNames[cfg.profile] || "Profile " + (cfg.profile + 1)}</span
      >
    </div>

    <!-- {#key cfg.profile} forces all inputs to be recreated when the active profile changes,
       preventing stale browser input state from leaking across profile switches. -->
    {#key cfg.profile}
      <!-- Wavelog section -->
      <section
        class="bg-surface-section border border-stroke-section rounded px-2.5 py-2"
      >
        <div
          class="text-2xs text-fg-muted uppercase tracking-wider mb-1.5 border-b border-stroke-section pb-1"
        >
          Wavelog
        </div>

        <div class="flex items-center gap-1.5 mb-1">
          <label
            class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
            for="wl-url">URL</label
          >
          <input
            id="wl-url"
            type="text"
            class="flex-1 w-full"
            value={activeProfile().wavelog_url}
            on:change={(e) => setProfileField("wavelog_url", e.target.value)}
            on:blur={loadStations}
            placeholder="https://log.example.com/index.php"
          />
        </div>

        <div class="flex items-center gap-1.5 mb-1">
          <label
            class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
            for="wl-key">API Key</label
          >
          <input
            id="wl-key"
            type="text"
            class="flex-1 w-full"
            value={activeProfile().wavelog_key}
            on:change={(e) => setProfileField("wavelog_key", e.target.value)}
            on:blur={loadStations}
          />
        </div>

        <div class="flex items-center gap-1.5 mb-1">
          <label
            class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
            for="wl-station">Station</label
          >
          <select
            id="wl-station"
            class="flex-1 w-full"
            on:change={(e) => setProfileField("wavelog_id", e.target.value)}
          >
            <option
              value="0"
              selected={activeProfile().wavelog_id === "0" ||
                activeProfile().wavelog_id === 0}>— select —</option
            >
            {#each stations as s}
              <option
                value={s.station_id}
                selected={String(s.station_id) ===
                  String(activeProfile().wavelog_id)}
              >
                {s.station_callsign} ({s.station_profile_name})
              </option>
            {/each}
          </select>
          <button
            class="py-0.5 px-1.5 text-sm"
            on:click={loadStations}
            title="Reload stations">↻</button
          >
        </div>

        <div class="flex items-center gap-1.5 mb-1">
          <label
            class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
            for="wl-radio">Radio name</label
          >
          <input
            id="wl-radio"
            type="text"
            class="flex-none w-field-sm"
            value={activeProfile().wavelog_radioname}
            on:change={(e) =>
              setProfileField("wavelog_radioname", e.target.value)}
          />
        </div>
      </section>

      <!-- Radio section -->
      <section
        class="bg-surface-section border border-stroke-section rounded px-2.5 py-2"
      >
        <div
          class="text-2xs text-fg-muted uppercase tracking-wider mb-1.5 border-b border-stroke-section pb-1"
        >
          Radio Control
        </div>

        <div class="flex items-center gap-1.5 mb-1">
          <label
            class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
            for="radio-type">Type</label
          >
          <select
            id="radio-type"
            class="flex-none w-field-sm"
            value={radioType}
            on:change={(e) => setRadioType(e.target.value)}
          >
            <option value="none">None</option>
            <option value="flrig">FLRig</option>
            <option value="hamlib">Hamlib</option>
          </select>
        </div>

        {#if radioType !== "none"}
          <div class="flex items-center gap-1.5 mb-1">
            <label
              class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
              for="radio-host">Host</label
            >
            <input
              id="radio-host"
              type="text"
              class="flex-none w-field-sm"
              value={radioType === "flrig"
                ? activeProfile().flrig_host
                : activeProfile().hamlib_host}
              on:change={(e) =>
                setProfileField(
                  radioType === "flrig" ? "flrig_host" : "hamlib_host",
                  e.target.value,
                )}
            />
            <label
              class="text-fg-label text-2xs ml-2 cursor-default"
              for="radio-port">Port</label
            >
            <input
              id="radio-port"
              type="text"
              class="flex-none w-field-xs"
              value={radioType === "flrig"
                ? activeProfile().flrig_port
                : activeProfile().hamlib_port}
              on:change={(e) =>
                setProfileField(
                  radioType === "flrig" ? "flrig_port" : "hamlib_port",
                  e.target.value,
                )}
            />
          </div>

          <div class="flex items-center gap-4 flex-wrap mb-1">
            <label>
              <input
                type="checkbox"
                checked={activeProfile().wavelog_pmode}
                on:change={(e) =>
                  setProfileField("wavelog_pmode", e.target.checked)}
              />
              Set MODE on QSY
            </label>
            {#if radioType === "hamlib"}
              <label>
                <input
                  type="checkbox"
                  checked={activeProfile().ignore_pwr}
                  on:change={(e) =>
                    setProfileField("ignore_pwr", e.target.checked)}
                />
                Ignore Power
              </label>
            {/if}
          </div>
        {/if}
      </section>
    {/key}

    <!-- Bottom buttons -->
    <div class="mt-1 flex flex-col items-center gap-1">
      <div class="flex gap-1.5 flex-wrap">
        <button on:click={save}>💾 Save</button>
        <button on:click={openProfileModal}>Profiles</button>
        <button on:click={test}>Test</button>
        <button on:click={openAdvancedModal}>⚙ Advanced</button>
        <button on:click={() => (showRotatorModal = true)}>Rot</button>
        <button on:click={quit}>Quit</button>
      </div>

      {#if saveMsg}
        <div class="alert alert-success text-2xs py-0.5 px-2 inline-block">
          {saveMsg}
        </div>
      {/if}
      {#if testMsg}
        <div
          class="alert inline-block py-0.5 px-2"
          class:alert-success={testSuccess}
          class:alert-danger={testSuccess === false}
          class:alert-info={testSuccess === null}
        >
          {testMsg}
        </div>
      {/if}
    </div>
  </div>
{/if}

<!-- Profile Modal -->
{#if showProfileModal}
  <div
    class="modal-overlay"
    on:click|self={() => (showProfileModal = false)}
    on:keydown={(e) => e.key === "Escape" && (showProfileModal = false)}
    role="dialog"
    aria-modal="true"
  >
    <div class="modal">
      <h4>Profiles</h4>

      <div class="flex flex-col gap-1 mb-2.5 max-h-48 overflow-y-auto">
        {#each cfg.profileNames as name, i}
          <div
            class="flex items-center justify-between px-2 py-1 bg-surface-section rounded border gap-1.5
              {i === cfg.profile
              ? 'border-stroke-accent'
              : 'border-stroke-section'}"
          >
            {#if renameIndex === i}
              <input
                type="text"
                bind:value={renameName}
                class="flex-1 bg-surface-input border border-stroke-base text-fg-base px-1.5 rounded text-xs"
              />
              <button on:click={doRenameProfile}>OK</button>
              <button on:click={() => (renameIndex = -1)}>✕</button>
            {:else}
              <span class="flex-1 text-xs">{name}</span>
              <div class="flex gap-1">
                {#if i !== cfg.profile}
                  <button on:click={() => doSwitchProfile(i)}>Switch</button>
                {:else}
                  <span class="text-xs text-accent-value px-1.5 py-px"
                    >Active</span
                  >
                {/if}
                <button
                  on:click={() => {
                    renameIndex = i;
                    renameName = name;
                  }}>Rename</button
                >
                {#if cfg.profileNames.length > 2 && i !== cfg.profile}
                  <button on:click={() => doDeleteProfile(i)}>Delete</button>
                {/if}
              </div>
            {/if}
          </div>
        {/each}
      </div>

      <div class="flex gap-1.5">
        <input
          type="text"
          bind:value={newProfileName}
          placeholder="New profile name"
          class="flex-1 bg-surface-input border border-stroke-base text-fg-base px-1.5 rounded text-xs"
        />
        <button on:click={doCreateProfile}>+ Add</button>
      </div>

      <div class="mt-2.5 text-right">
        <button on:click={() => (showProfileModal = false)}>Close</button>
      </div>
    </div>
  </div>
{/if}

<!-- Advanced Modal -->
{#if showAdvancedModal}
  <div
    class="modal-overlay"
    on:click|self={() => (showAdvancedModal = false)}
    on:keydown={(e) => e.key === "Escape" && (showAdvancedModal = false)}
    role="dialog"
    aria-modal="true"
  >
    <div class="modal">
      <h4>Advanced Settings</h4>

      <div class="flex items-center gap-1.5 mb-1">
        <label>
          <input type="checkbox" bind:checked={advUdpEnabled} />
          UDP Listener enabled
        </label>
      </div>

      <div class="flex items-center gap-1.5 mb-1">
        <label
          class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
          for="adv-port">UDP Port</label
        >
        <input
          id="adv-port"
          type="number"
          class="flex-none w-field-xs"
          bind:value={advUdpPort}
          min="1024"
          max="65535"
          disabled={!advUdpEnabled}
        />
      </div>

      {#if advStatus}
        <div class="alert alert-info mt-2">{advStatus}</div>
      {/if}

      <div class="mt-3 flex gap-1.5 justify-end">
        <button on:click={doSaveAdvanced}>Save</button>
        <button on:click={() => (showAdvancedModal = false)}>Cancel</button>
      </div>
    </div>
  </div>
{/if}

<!-- Rotator Modal -->
{#if showRotatorModal && cfg}
  <div
    class="modal-overlay"
    on:click|self={() => (showRotatorModal = false)}
    on:keydown={(e) => e.key === "Escape" && (showRotatorModal = false)}
    role="dialog"
    aria-modal="true"
  >
    <div class="modal">
      <h4>Rotator Settings</h4>

      <div class="flex items-center gap-1.5 mb-1">
        <label
          class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
          for="rot-host">Host</label
        >
        <input
          id="rot-host"
          type="text"
          class="flex-none w-field-sm"
          value={activeProfile().rotator_host}
          on:change={(e) => setProfileField("rotator_host", e.target.value)}
          placeholder="leave empty to disable"
        />
      </div>

      <div class="flex items-center gap-1.5 mb-1">
        <label
          class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
          for="rot-port">Port</label
        >
        <input
          id="rot-port"
          type="text"
          class="flex-none w-field-xs"
          value={activeProfile().rotator_port}
          on:change={(e) => setProfileField("rotator_port", e.target.value)}
        />
      </div>

      <div class="flex items-center gap-1.5 mb-1">
        <label
          class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
          >Threshold Az</label
        >
        <input
          type="number"
          class="flex-none w-field-xs"
          value={activeProfile().rotator_threshold_az}
          on:change={(e) =>
            setProfileField("rotator_threshold_az", Number(e.target.value))}
          min="0"
          max="360"
          step="0.5"
        />
        <span class="text-fg-muted text-2xs">°</span>
        <label class="text-fg-label text-2xs ml-2 cursor-default">El</label>
        <input
          type="number"
          class="flex-none w-field-xs"
          value={activeProfile().rotator_threshold_el}
          on:change={(e) =>
            setProfileField("rotator_threshold_el", Number(e.target.value))}
          min="0"
          max="90"
          step="0.5"
        />
        <span class="text-fg-muted text-2xs">°</span>
      </div>

      <div class="flex items-center gap-1.5 mb-1">
        <label
          class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end"
          >Park Az</label
        >
        <input
          type="number"
          class="flex-none w-field-xs"
          value={activeProfile().rotator_park_az}
          on:change={(e) =>
            setProfileField("rotator_park_az", Number(e.target.value))}
          min="0"
          max="360"
          step="1"
        />
        <span class="text-fg-muted text-2xs">°</span>
        <label class="text-fg-label text-2xs ml-2 cursor-default">El</label>
        <input
          type="number"
          class="flex-none w-field-xs"
          value={activeProfile().rotator_park_el}
          on:change={(e) =>
            setProfileField("rotator_park_el", Number(e.target.value))}
          min="0"
          max="360"
          step="1"
        />
        <span class="text-fg-muted text-2xs">°</span>
      </div>

      <div class="mt-3 text-right">
        <button on:click={() => (showRotatorModal = false)}>Close</button>
      </div>
    </div>
  </div>
{/if}
