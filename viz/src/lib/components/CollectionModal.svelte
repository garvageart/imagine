<script lang="ts">
    import type { EventHandler } from "svelte/elements";
    import Button from "./Button.svelte";
    import ModalOverlay from "./modal/ModalOverlay.svelte";
    import SliderToggle from "./SliderToggle.svelte";
    import InputText from "./dom/InputText.svelte";

    interface Props {
        heading: string;
        data: {
            name: string;
            description?: string;
            isPrivate?: boolean | null;
        };
        buttonText: string;
        modalAction:
            | EventHandler<SubmitEvent, HTMLFormElement>
            | null
            | undefined;
    }

    let {
        heading,
        data = $bindable(),
        buttonText,
        modalAction,
    }: Props = $props();

    let allData = $state({ ...data, isPrivate: data.isPrivate ? "on" : "off" });
</script>

<ModalOverlay>
    <div id="viz-collection-modal">
        <h1>{heading}</h1>
        <form id="collection-form" onsubmit={modalAction}>
            <InputText
                id="collection-name"
                name="name"
                label="Name"
                placeholder="Name"
                type="text"
                bind:value={allData.name}
                required
                autocomplete="on"
            />
            <!-- svelte-ignore element_invalid_self_closing_tag -->
            <textarea
                id="collection-description"
                name="description"
                placeholder="Description (optional)"
                bind:value={allData.description}
            />
            <SliderToggle
                id="collection-private"
                style="margin-bottom: 1rem;"
                label="Private"
                bind:value={allData.isPrivate as "on" | "off"}
            />
            <Button style="margin-top: 1rem;">
                <input
                    id="collection-submit"
                    type="submit"
                    value={buttonText}
                />
            </Button>
        </form>
    </div>
</ModalOverlay>

<style lang="scss">
    #viz-collection-modal {
        width: 100%;
        height: 100%;
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        align-items: center;
    }

    form {
        width: 60%;
        display: flex;
        flex-direction: column;
        justify-content: flex-start;
        align-items: flex-start;
        height: 80%;
        max-height: 100%;
        margin-top: 1em;
        position: relative;
    }

    textarea {
        width: 100%;
        max-width: 100%;
        min-width: 100%;
        min-height: 2rem;
        color: var(--imag-text-color);
        background-color: var(--imag-bg-color);
        outline: none;
        border: none;
        box-shadow: 0 -1.5px 0 var(--imag-60) inset;
        font-size: 2rem;
        font-family: var(--imag-font-family);
        font-weight: bold;
        padding: 0.5rem 1rem;
        margin-bottom: 1em;

        &::placeholder {
            color: var(--imag-40);
            font-family: var(--imag-font-family);
        }

        &:focus::placeholder {
            color: var(--imag-60);
        }

        &:focus {
            background-color: var(--imag-100);
            box-shadow: 0 -2px 0 var(--imag-primary) inset;
        }

        &:-webkit-autofill,
        &:-webkit-autofill:focus {
            -webkit-text-fill-color: var(--imag-text-color);
            -webkit-box-shadow: 0 0 0px 1000px var(--imag-100) inset;
            -webkit-box-shadow: 0 -5px 0 var(--imag-primary) inset;
            transition:
                background-color 0s 600000s,
                color 0s 600000s !important;
        }
    }

    #collection-description {
        font-size: 1rem;
        resize: none;
        font-weight: 400;
    }

    #collection-submit {
        border: inherit;
        background-color: transparent;
        color: inherit;
        font-family: inherit;
        font-weight: bold;
        font-size: inherit;
        cursor: pointer;
        width: 100%;
        height: 100%;
    }
</style>
