<script lang="ts">
    import Checkbox from './Checkbox.svelte';

    interface Option {
        label: string;
        value: string;
    }

    interface Props {
        options: Option[];
        value: string[];
    }

    let { options, value = $bindable() }: Props = $props();

    function updateValue(optionValue: string, checked: boolean) {
        if (checked) {
            if (!value.includes(optionValue)) {
                value = [...value, optionValue];
            }
        } else {
            value = value.filter(v => v !== optionValue);
        }
    }
</script>

<div class="checkbox-group">
    {#each options as option (option.value)}
        <div class="checkbox-group-item">
            <Checkbox
                label={option.label}
                checked={value.includes(option.value)}
                onchange={(e: Event) => updateValue(option.value, (e.target as HTMLInputElement).checked)}
            />
        </div>
    {/each}
</div>

<style>
    .checkbox-group {
        display: flex;
        flex-direction: column;
        gap: 0.5rem;
    }
    
    .checkbox-group-item {
        display: flex;
    }
</style>
