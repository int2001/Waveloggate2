<script>
  import { createEventDispatcher } from "svelte";
  const dispatch = createEventDispatcher();

  export let profile;
</script>

<div
  class="modal-overlay"
  on:click|self={() => dispatch("close")}
  on:keydown={(e) => e.key === "Escape" && dispatch("close")}
  role="dialog"
  aria-modal="true"
>
  <div class="modal">
    <h4>Rotator Settings</h4>

    <div class="flex items-center gap-1.5 mb-1">
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="rot-host">Host</label>
      <input
        id="rot-host"
        type="text"
        class="flex-none w-field-sm"
        value={profile.rotator_host}
        on:change={(e) => dispatch("fieldchange", { key: "rotator_host", value: e.target.value })}
        placeholder="leave empty to disable"
      />
    </div>

    <div class="flex items-center gap-1.5 mb-1">
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="rot-port">Port</label>
      <input
        id="rot-port"
        type="text"
        class="flex-none w-field-xs"
        value={profile.rotator_port}
        on:change={(e) => dispatch("fieldchange", { key: "rotator_port", value: e.target.value })}
      />
    </div>

    <div class="flex items-center gap-1.5 mb-1">
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end">Threshold Az</label>
      <input
        type="number"
        class="flex-none w-field-xs"
        value={profile.rotator_threshold_az}
        on:change={(e) => dispatch("fieldchange", { key: "rotator_threshold_az", value: Number(e.target.value) })}
        min="0" max="360" step="0.5"
      />
      <span class="text-fg-muted text-2xs">°</span>
      <label class="text-fg-label text-2xs ml-2 cursor-default">El</label>
      <input
        type="number"
        class="flex-none w-field-xs"
        value={profile.rotator_threshold_el}
        on:change={(e) => dispatch("fieldchange", { key: "rotator_threshold_el", value: Number(e.target.value) })}
        min="0" max="90" step="0.5"
      />
      <span class="text-fg-muted text-2xs">°</span>
    </div>

    <div class="flex items-center gap-1.5 mb-1">
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end">Park Az</label>
      <input
        type="number"
        class="flex-none w-field-xs"
        value={profile.rotator_park_az}
        on:change={(e) => dispatch("fieldchange", { key: "rotator_park_az", value: Number(e.target.value) })}
        min="0" max="360" step="1"
      />
      <span class="text-fg-muted text-2xs">°</span>
      <label class="text-fg-label text-2xs ml-2 cursor-default">El</label>
      <input
        type="number"
        class="flex-none w-field-xs"
        value={profile.rotator_park_el}
        on:change={(e) => dispatch("fieldchange", { key: "rotator_park_el", value: Number(e.target.value) })}
        min="0" max="360" step="1"
      />
      <span class="text-fg-muted text-2xs">°</span>
    </div>

    <div class="mt-3 text-right">
      <button on:click={() => dispatch("close")}>Close</button>
    </div>
  </div>
</div>
