<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let profile;
  export let radioType = "none";
</script>

<section class="bg-surface-section border border-stroke-section rounded-lg px-4 py-3">
  <div class="text-2xs text-fg-bright font-semibold uppercase tracking-wider mb-3 pl-2 border-l-2 border-stroke-accent">
    Radio Control
  </div>

  <div class="flex items-center gap-2 mb-1.5">
    <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs" for="radio-type">Type</label>
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
    <div class="flex items-center gap-2 mb-1.5">
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs" for="radio-host">Host</label>
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
      <label class="text-fg-label text-2xs ml-1 cursor-default" for="radio-port">Port</label>
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

    <div class="flex items-center gap-5 flex-wrap">
      <label class="text-fg-label text-xs">
        <input
          type="checkbox"
          checked={profile.wavelog_pmode}
          on:change={(e) => dispatch("fieldchange", { key: "wavelog_pmode", value: e.target.checked })}
        />
        Set MODE on QSY
      </label>
      {#if radioType === "hamlib"}
        <label class="text-fg-label text-xs">
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
