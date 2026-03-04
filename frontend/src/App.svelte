<script lang="ts">
  import {Translate, GetModels, SwitchModel} from '../wailsjs/go/main/App.js'
  import { onMount } from 'svelte';

  let query: string = ""
  let aiResult: string = ""
  let isLoading: boolean = false
  let models: string[] = []
  let selectedModel: string = ""

  onMount(() => {
     GetModels().then(res => {
         if (res && res.length > 0) {
             models = res;
             selectedModel = models[0]; // pick first model by default
         }
     });
  });

  function search(): void {
    const q = query.trim()
    if (!q) {
       aiResult = "";
       return;
    }
    
    isLoading = true
    Translate(q).then(res => {
        isLoading = false
        if (res) {
            aiResult = res;
        } else {
            aiResult = "";
        }
    }).catch(err => {
        console.error(err);
        isLoading = false
        aiResult = "";
    })
  }

  function handleModelChange(): void {
      if (selectedModel) {
          aiResult = "Loading...";
          SwitchModel(selectedModel).then(name => {
              if (name) {
                  aiResult = "🔄 Switch to " + name + " successfully. Try translating something!";
              } else {
                  aiResult = "Failed to switch model";
              }
          });
      }
  }
</script>

<main>
  <h2>AI Wails Translator</h2>
  <div class="model-picker">
    <span class="model-badge">Model:</span>
    <select bind:value={selectedModel} on:change={handleModelChange} class="model-select">
       {#each models as model}
           <option value={model}>{model}</option>
       {/each}
    </select>
  </div>
  
  <div class="search-box">
    <input autocomplete="off" bind:value={query} placeholder="Enter Chinese or English sentence..." class="input" type="text" on:keydown={e => e.key === 'Enter' && search()}/>
    <button class="btn" on:click={search}>Translate</button>
  </div>
  
  <div class="results">
      {#if isLoading}
          <div class="loader">Loading AI Translation...</div>
      {/if}

      {#if aiResult}
          <div class="ai-translation-card">
              <h3>✨ AI Contextual Translation</h3>
              <p>{aiResult}</p>
          </div>
      {/if}
  </div>
</main>

<style>
  main {
      padding: 20px;
      font-family: Arial, sans-serif;
  }
  h2 {
      text-align: center;
  }
  .model-picker {
      display: flex;
      justify-content: center;
      align-items: center;
      gap: 15px;
      margin-bottom: 25px;
  }
  .model-badge {
      background: #eee;
      padding: 5px 12px;
      border-radius: 12px;
      font-size: 0.9em;
      color: #555;
  }
  .model-select {
      padding: 8px 15px;
      font-size: 0.9em;
      border-radius: 4px;
      border: 1px solid #ccc;
  }
  .search-box {
      display: flex;
      justify-content: center;
      margin-bottom: 20px;
  }
  .input {
      width: 400px;
      padding: 10px;
      font-size: 16px;
      border: 1px solid #ccc;
      border-radius: 4px;
      margin-right: 10px;
  }
  .btn {
      padding: 10px 20px;
      font-size: 16px;
      cursor: pointer;
      background-color: #007bff;
      color: white;
      border: none;
      border-radius: 4px;
  }
  .results {
      max-width: 600px;
      margin: 0 auto;
      text-align: left;
  }
  .ai-translation-card {
      background: linear-gradient(135deg, #f6d365 0%, #fda085 100%);
      color: #333;
      padding: 20px;
      border-radius: 10px;
      margin-bottom: 20px;
      box-shadow: 0 4px 6px rgba(0,0,0,0.1);
  }
  .ai-translation-card h3 {
      margin-top: 0;
      color: #7a3e00;
  }
  .ai-translation-card p {
      font-size: 1.2em;
      line-height: 1.5;
  }
  .loader {
      text-align: center;
      padding: 20px;
      font-style: italic;
      color: #007bff;
      animation: pulse 1.5s infinite;
  }
  @keyframes pulse {
      0% { opacity: 0.5; }
      50% { opacity: 1; }
      100% { opacity: 0.5; }
  }
</style>
