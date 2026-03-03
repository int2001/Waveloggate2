<script>
  import { onMount, createEventDispatcher } from "svelte";
  import { GetUDPStatus, SaveAdvanced } from "../../../wailsjs/go/main/App.js";

  const dispatch = createEventDispatcher();

  let advUdpEnabled = true;
  let advUdpPort = 2333;
  let advStatus = "";

  onMount(async () => {
    const status = await GetUDPStatus();
    advUdpEnabled = status.enabled;
    advUdpPort = status.port;
  });

  async function save() {
    const err = await SaveAdvanced(advUdpEnabled, advUdpPort);
    if (err) {
      advStatus = "Error: " + err;
    } else {
      advStatus = "Saved ✓";
      setTimeout(() => {
        advStatus = "";
        dispatch("close");
      }, 1500);
    }
  }
</script>

<div
  class="modal-overlay"
  on:click|self={() => dispatch("close")}
  on:keydown={(e) => e.key === "Escape" && dispatch("close")}
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
      <label class="w-field-xs flex-shrink-0 text-fg-label text-2xs justify-end" for="adv-port">UDP Port</label>
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
      <button on:click={save}>Save</button>
      <button on:click={() => dispatch("close")}>Cancel</button>
    </div>
  </div>
</div>
