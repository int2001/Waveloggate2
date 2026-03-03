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
  <div class="loading">Loading…</div>
{:else if cfg}
  <div class="config-tab">
    <!-- Profile indicator -->
    <div class="profile-bar">
      <span class="profile-label">Profile:</span>
      <span class="profile-name"
        >{cfg.profileNames[cfg.profile] || "Profile " + (cfg.profile + 1)}</span
      >
    </div>

    <!-- {#key cfg.profile} forces all inputs to be recreated when the active profile changes,
       preventing stale browser input state from leaking across profile switches. -->
    {#key cfg.profile}
      <!-- Wavelog section -->
      <section>
        <div class="section-title">Wavelog</div>

        <div class="row">
          <label class="field-label" for="wl-url">URL</label>
          <input
            id="wl-url"
            type="text"
            class="field-input"
            value={activeProfile().wavelog_url}
            on:change={(e) => setProfileField("wavelog_url", e.target.value)}
            on:blur={loadStations}
            placeholder="https://log.example.com/index.php"
          />
        </div>

        <div class="row">
          <label class="field-label" for="wl-key">API Key</label>
          <input
            id="wl-key"
            type="text"
            class="field-input"
            value={activeProfile().wavelog_key}
            on:change={(e) => setProfileField("wavelog_key", e.target.value)}
            on:blur={loadStations}
          />
        </div>

        <div class="row">
          <label class="field-label" for="wl-station">Station</label>
          <select
            id="wl-station"
            class="field-input"
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
            class="icon-btn"
            on:click={loadStations}
            title="Reload stations">↻</button
          >
        </div>

        <div class="row">
          <label class="field-label" for="wl-radio">Radio name</label>
          <input
            id="wl-radio"
            type="text"
            class="field-input short"
            value={activeProfile().wavelog_radioname}
            on:change={(e) =>
              setProfileField("wavelog_radioname", e.target.value)}
          />
        </div>
      </section>

      <!-- Radio section -->
      <section>
        <div class="section-title">Radio Control</div>

        <div class="row">
          <label class="field-label" for="radio-type">Type</label>
          <select
            id="radio-type"
            class="field-input short"
            value={radioType}
            on:change={(e) => setRadioType(e.target.value)}
          >
            <option value="none">None</option>
            <option value="flrig">FLRig</option>
            <option value="hamlib">Hamlib</option>
          </select>
        </div>

        {#if radioType !== "none"}
          <div class="row">
            <label class="field-label" for="radio-host">Host</label>
            <input
              id="radio-host"
              type="text"
              class="field-input short"
              value={radioType === "flrig"
                ? activeProfile().flrig_host
                : activeProfile().hamlib_host}
              on:change={(e) =>
                setProfileField(
                  radioType === "flrig" ? "flrig_host" : "hamlib_host",
                  e.target.value,
                )}
            />
            <label class="field-label" for="radio-port" style="margin-left:8px"
              >Port</label
            >
            <input
              id="radio-port"
              type="text"
              class="field-input xshort"
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

          <div class="row checkrow">
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
    <div class="bottom-bar">
      <div class="btn-row">
        <button on:click={save}>💾 Save</button>
        <button on:click={openProfileModal}>Profiles</button>
        <button on:click={test}>Test</button>
        <button on:click={openAdvancedModal}>⚙ Advanced</button>
        <button on:click={() => (showRotatorModal = true)}>Rot</button>
        <button on:click={quit}>Quit</button>
      </div>

      {#if saveMsg}
        <div class="alert alert-success inline-msg">{saveMsg}</div>
      {/if}
      {#if testMsg}
        <div
          class="alert"
          class:alert-success={testSuccess}
          class:alert-danger={testSuccess === false}
          class:alert-info={testSuccess === null}
          style="display:inline-block;padding:2px 8px"
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

      <div class="profile-list">
        {#each cfg.profileNames as name, i}
          <div class="profile-item" class:active-profile={i === cfg.profile}>
            {#if renameIndex === i}
              <input type="text" bind:value={renameName} class="rename-input" />
              <button on:click={doRenameProfile}>OK</button>
              <button on:click={() => (renameIndex = -1)}>✕</button>
            {:else}
              <span class="profile-item-name">{name}</span>
              <div class="profile-item-btns">
                {#if i !== cfg.profile}
                  <button on:click={() => doSwitchProfile(i)}>Switch</button>
                {:else}
                  <span class="active-badge">Active</span>
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

      <div class="new-profile-row">
        <input
          type="text"
          bind:value={newProfileName}
          placeholder="New profile name"
          class="rename-input"
        />
        <button on:click={doCreateProfile}>+ Add</button>
      </div>

      <div style="margin-top:10px;text-align:right">
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

      <div class="row">
        <label>
          <input type="checkbox" bind:checked={advUdpEnabled} />
          UDP Listener enabled
        </label>
      </div>

      <div class="row">
        <label class="field-label" for="adv-port">UDP Port</label>
        <input
          id="adv-port"
          type="number"
          class="field-input xshort"
          bind:value={advUdpPort}
          min="1024"
          max="65535"
          disabled={!advUdpEnabled}
        />
      </div>

      {#if advStatus}
        <div class="alert alert-info" style="margin-top:8px">{advStatus}</div>
      {/if}

      <div
        style="margin-top:12px;text-align:right;display:flex;gap:6px;justify-content:flex-end"
      >
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

      <div class="row">
        <label class="field-label" for="rot-host">Host</label>
        <input
          id="rot-host"
          type="text"
          class="field-input short"
          value={activeProfile().rotator_host}
          on:change={(e) => setProfileField("rotator_host", e.target.value)}
          placeholder="leave empty to disable"
        />
      </div>

      <div class="row">
        <label class="field-label" for="rot-port">Port</label>
        <input
          id="rot-port"
          type="text"
          class="field-input xshort"
          value={activeProfile().rotator_port}
          on:change={(e) => setProfileField("rotator_port", e.target.value)}
        />
      </div>

      <div class="row">
        <label class="field-label">Threshold Az</label>
        <input
          type="number"
          class="field-input xshort"
          value={activeProfile().rotator_threshold_az}
          on:change={(e) =>
            setProfileField("rotator_threshold_az", Number(e.target.value))}
          min="0"
          max="360"
          step="0.5"
        />
        <span class="unit">°</span>
        <label class="field-label" style="margin-left:8px">El</label>
        <input
          type="number"
          class="field-input xshort"
          value={activeProfile().rotator_threshold_el}
          on:change={(e) =>
            setProfileField("rotator_threshold_el", Number(e.target.value))}
          min="0"
          max="90"
          step="0.5"
        />
        <span class="unit">°</span>
      </div>

      <div class="row">
        <label class="field-label">Park Az</label>
        <input
          type="number"
          class="field-input xshort"
          value={activeProfile().rotator_park_az}
          on:change={(e) =>
            setProfileField("rotator_park_az", Number(e.target.value))}
          min="0"
          max="360"
          step="1"
        />
        <span class="unit">°</span>
        <label class="field-label" style="margin-left:8px">El</label>
        <input
          type="number"
          class="field-input xshort"
          value={activeProfile().rotator_park_el}
          on:change={(e) =>
            setProfileField("rotator_park_el", Number(e.target.value))}
          min="0"
          max="90"
          step="1"
        />
        <span class="unit">°</span>
      </div>

      <div style="margin-top:12px;text-align:right">
        <button on:click={() => (showRotatorModal = false)}>Close</button>
      </div>
    </div>
  </div>
{/if}

<style>
  .config-tab {
    padding: 8px 12px;
    display: flex;
    flex-direction: column;
    gap: 6px;
  }

  .loading {
    padding: 20px;
    color: #888;
    text-align: center;
  }

  .profile-bar {
    display: flex;
    align-items: center;
    gap: 6px;
    background: #262626;
    padding: 3px 8px;
    border-radius: 3px;
    font-size: 11px;
  }

  .profile-label {
    color: #888;
  }
  .profile-name {
    color: #5af;
  }

  section {
    background: #2a2a2a;
    border: 1px solid #404040;
    border-radius: 4px;
    padding: 8px 10px;
  }

  .section-title {
    font-size: 11px;
    color: #888;
    text-transform: uppercase;
    letter-spacing: 0.5px;
    margin-bottom: 6px;
    border-bottom: 1px solid #404040;
    padding-bottom: 4px;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 6px;
    margin-bottom: 5px;
  }

  .checkrow {
    gap: 16px;
    flex-wrap: wrap;
  }

  .field-label {
    width: 70px;
    flex-shrink: 0;
    color: #aaa;
    font-size: 11px;
    text-align: right;
  }

  .field-input {
    flex: 1;
    width: 100%;
  }

  .field-input.short {
    flex: 0;
    width: 120px;
  }
  .field-input.xshort {
    flex: 0;
    width: 70px;
  }

  .unit {
    color: #888;
    font-size: 11px;
  }

  .icon-btn {
    padding: 2px 6px;
    font-size: 14px;
  }

  .bottom-bar {
    margin-top: 4px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
  }

  .btn-row {
    display: flex;
    gap: 6px;
    flex-wrap: wrap;
  }

  .inline-msg {
    font-size: 11px;
    padding: 2px 8px;
    display: inline-block;
  }

  /* Profile modal internals */
  .profile-list {
    display: flex;
    flex-direction: column;
    gap: 4px;
    margin-bottom: 10px;
    max-height: 200px;
    overflow-y: auto;
  }

  .profile-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 4px 8px;
    background: #2a2a2a;
    border-radius: 3px;
    border: 1px solid #404040;
    gap: 6px;
  }

  .profile-item.active-profile {
    border-color: #5a9fd4;
  }

  .profile-item-name {
    flex: 1;
    font-size: 12px;
  }

  .profile-item-btns {
    display: flex;
    gap: 4px;
  }

  .active-badge {
    font-size: 10px;
    color: #5af;
    padding: 1px 6px;
  }

  .new-profile-row {
    display: flex;
    gap: 6px;
  }

  .rename-input {
    flex: 1;
    background: #404040;
    border: 1px solid #555;
    color: #c6c6c6;
    padding: 3px 6px;
    border-radius: 3px;
    font-size: 12px;
  }
</style>
