<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let profile;
  export let radioType = "none";

  // Track last active type so re-enabling restores the previous selection.
  let lastType = radioType !== "none" ? radioType : "flrig";
  $: if (radioType !== "none") lastType = radioType;

  $: radioEnabled = radioType !== "none";

  function onEnableChange(e) {
    dispatch("typechange", e.currentTarget.checked ? lastType : "none");
  }

  function onTypeChange(e) {
    dispatch("typechange", e.currentTarget.value);
  }
</script>

<section class="bg-surface-section border border-stroke-section rounded-lg px-4 py-3">
  <div class="flex items-center justify-between {radioEnabled ? 'mb-3' : ''}">
    <div class="text-2xs text-fg-bright font-semibold uppercase tracking-wider pl-2 border-l-2 border-stroke-accent">
      Radio Control
    </div>
    <label class="text-fg-label text-xs">
      <input
        type="checkbox"
        checked={radioEnabled}
        on:change={onEnableChange}
      />
      Enable
    </label>
  </div>

  {#if radioEnabled}
    <div class="flex flex-col gap-1.5">
      <div class="flex items-center gap-2">
        <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs" for="radio-type">Type</label>
        <select
          id="radio-type"
          class="flex-none w-field-sm"
          value={radioType}
          on:change={onTypeChange}
        >
          <option value="flrig">FLRig</option>
          <option value="hamlib">Hamlib</option>
        </select>
      </div>

      <div class="flex items-center gap-2">
        <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs" for="radio-host">Host</label>
        <input
          id="radio-host"
          type="text"
          class="flex-none w-field-sm"
          value={radioType === "flrig" ? profile.flrig_host : profile.hamlib_host}
          on:change={(e) => dispatch("fieldchange", {
            key: radioType === "flrig" ? "flrig_host" : "hamlib_host",
            value: e.currentTarget.value,
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
            value: e.currentTarget.value,
          })}
        />
      </div>

      <div class="flex items-center gap-5 flex-wrap">
        <label class="text-fg-label text-xs">
          <input
            type="checkbox"
            checked={profile.wavelog_pmode}
            on:change={(e) => dispatch("fieldchange", { key: "wavelog_pmode", value: e.currentTarget.checked })}
          />
          Set MODE on QSY
        </label>
        {#if radioType === "hamlib"}
          <label class="text-fg-label text-xs">
            <input
              type="checkbox"
              checked={profile.ignore_pwr}
              on:change={(e) => dispatch("fieldchange", { key: "ignore_pwr", value: e.currentTarget.checked })}
            />
            Ignore Power
          </label>
        {/if}
      </div>
    </div>
  {/if}
</section>
