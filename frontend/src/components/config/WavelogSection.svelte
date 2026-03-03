<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let profile;
  export let stations = [];
</script>

<section class="bg-surface-section border border-stroke-section rounded px-2.5 py-2">
  <div class="text-2xs text-fg-muted font-medium uppercase tracking-wider mb-1.5 border-b border-stroke-section pb-1">
    Wavelog
  </div>

  <div class="flex items-center gap-1.5 mb-1">
    <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="wl-url">URL</label>
    <input
      id="wl-url"
      type="text"
      class="flex-1 w-full"
      value={profile.wavelog_url}
      on:change={(e) => dispatch("fieldchange", { key: "wavelog_url", value: e.target.value })}
      on:blur={() => dispatch("reloadstations")}
      placeholder="https://log.example.com/index.php"
    />
  </div>

  <div class="flex items-center gap-1.5 mb-1">
    <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="wl-key">API Key</label>
    <input
      id="wl-key"
      type="text"
      class="flex-1 w-full"
      value={profile.wavelog_key}
      on:change={(e) => dispatch("fieldchange", { key: "wavelog_key", value: e.target.value })}
      on:blur={() => dispatch("reloadstations")}
    />
  </div>

  <div class="flex items-center gap-1.5 mb-1">
    <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="wl-station">Station</label>
    <select
      id="wl-station"
      class="flex-1 w-full"
      on:change={(e) => dispatch("fieldchange", { key: "wavelog_id", value: e.target.value })}
    >
      <option value="0" selected={profile.wavelog_id === "0" || profile.wavelog_id === 0}>— select —</option>
      {#each stations as s}
        <option value={s.station_id} selected={String(s.station_id) === String(profile.wavelog_id)}>
          {s.station_callsign} ({s.station_profile_name})
        </option>
      {/each}
    </select>
    <button class="py-0.5 px-1.5 text-sm" on:click={() => dispatch("reloadstations")} title="Reload stations">↻</button>
  </div>

  <div class="flex items-center gap-1.5 mb-1">
    <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="wl-radio">Radio name</label>
    <input
      id="wl-radio"
      type="text"
      class="flex-none w-field-sm"
      value={profile.wavelog_radioname}
      on:change={(e) => dispatch("fieldchange", { key: "wavelog_radioname", value: e.target.value })}
    />
  </div>
</section>
