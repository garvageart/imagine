<script lang="ts">
    import Checkbox from "$lib/components/dom/Checkbox.svelte";
    import Button from "$lib/components/Button.svelte";
	import InputText from "../dom/InputText.svelte";

    interface Props {
        title: string;
        items: Map<string, number>; // Value -> Count
        selected: string[];
        onChange: (selected: string[]) => void;
    }

    let { title, items, selected, onChange }: Props = $props();

    let searchTerm = $state("");
    let isExpanded = $state(false);

    // Convert map to array and sort by count desc
    let sortedItems = $derived(
        Array.from(items.entries())
            .sort((a, b) => b[1] - a[1])
            .map(([value, count]) => ({ value, count }))
    );

    let filteredItems = $derived(
        sortedItems.filter(item => 
            item.value.toLowerCase().includes(searchTerm.toLowerCase())
        )
    );

    let displayItems = $derived(
        isExpanded ? filteredItems : filteredItems.slice(0, 5)
    );

    function toggle(value: string) {
        if (selected.includes(value)) {
            onChange(selected.filter(s => s !== value));
        } else {
            onChange([...selected, value]);
        }
    }
</script>

<div class="facet-container">
    <div class="facet-list">
        {#each displayItems as item (item.value)}
            <div class="facet-item">
                <Checkbox 
                    label={item.value} 
                    checked={selected.includes(item.value)}
                    onchange={() => toggle(item.value)}
                />
                <span class="count">({item.count})</span>
            </div>
        {/each}
        
        {#if filteredItems.length === 0}
             <div class="empty">No items found</div>
        {/if}
    </div>

    {#if filteredItems.length > 5}
        <Button 
            variant="mini" 
            class="more-btn" 
            onclick={() => isExpanded = !isExpanded}
        >
            {isExpanded ? "Show Less" : `Show All (${filteredItems.length})`}
        </Button>
    {/if}
</div>

<style lang="scss">
    .facet-container {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    } 

    .facet-list {
        display: flex;
        flex-direction: column;
        gap: 4px;
    }

    .facet-item {
        display: flex;
        align-items: center;
        justify-content: space-between;
        font-size: 0.85rem;
        color: var(--imag-60);
        
        /* Align checkbox properly */
        :global(.checkbox-wrapper) {
            flex: 1;
            min-width: 0; /* allows text truncation if needed */
        }
        
        :global(.label-text) {
            font-size: 0.85rem;
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
        }
    }

    .count {
        font-size: 0.75rem;
        opacity: 0.7;
        margin-left: 8px;
    }
    
    .empty {
        font-style: italic;
        font-size: 0.8rem;
        opacity: 0.6;
        color: var(--imag-60);
    }

    /* Override Button styling to look like a link or simple toggle */
    :global(.more-btn) {
        background-color: transparent !important;
        color: var(--imag-primary) !important;
        padding: 0 !important;
        align-self: flex-start;
        height: auto !important;
        margin-top: 4px;
        
        &:hover {
            text-decoration: underline;
        }
    }
</style>