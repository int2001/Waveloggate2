<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let profile;
  export let radioType = "none";
</script>

<section class="bg-surface-section border border-stroke-section rounded px-2.5 py-2">
  <div class="text-2xs text-fg-muted font-medium uppercase tracking-wider mb-1.5 border-b border-stroke-section pb-1">
    Radio Control
  </div>

  <div class="flex items-center gap-1.5 mb-1">
    <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="radio-type">Type</label>
    <select
      id="radio-type"
      class="flex-none w-field-sm"
      value={radioType}
      on:change={(e) => dispatch("typechange", e.target.value)}
    >
      <option value="none">None</option>
      <option value="flrig">FLRig</option>
      <option value="hamlib">Hamlib</option>
    </select>
  </div>

  {#if radioType !== "none"}
    <div class="flex items-center gap-1.5 mb-1">
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="radio-host">Host</label>
      <input
        id="radio-host"
        type="text"
        class="flex-none w-field-sm"
        value={radioType === "flrig" ? profile.flrig_host : profile.hamlib_host}
        on:change={(e) => dispatch("fieldchange", {
          key: radioType === "flrig" ? "flrig_host" : "hamlib_host",
          value: e.target.value,
        })}
      />
      <label class="text-fg-label text-2xs ml-2 cursor-default" for="radio-port">Port</label>
      <input
        id="radio-port"
        type="text"
        class="flex-none w-field-xs"
        value={radioType === "flrig" ? profile.flrig_port : profile.hamlib_port}
        on:change={(e) => dispatch("fieldchange", {
          key: radioType === "flrig" ? "flrig_port" : "hamlib_port",
          value: e.target.value,
        })}
      />
    </div>

    <div class="flex items-center gap-4 flex-wrap mb-1">
      <label>
        <input
          type="checkbox"
          checked={profile.wavelog_pmode}
          on:change={(e) => dispatch("fieldchange", { key: "wavelog_pmode", value: e.target.checked })}
        />
        Set MODE on QSY
      </label>
      {#if radioType === "hamlib"}
        <label>
          <input
            type="checkbox"
            checked={profile.ignore_pwr}
            on:change={(e) => dispatch("fieldchange", { key: "ignore_pwr", value: e.target.checked })}
          />
          Ignore Power
        </label>
      {/if}
    </div>
  {/if}
</section>
