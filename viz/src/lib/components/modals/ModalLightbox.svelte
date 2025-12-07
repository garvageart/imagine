<script lang="ts">
	import { modal } from "$lib/states/index.svelte";
	import type { Snippet } from "svelte";
	import type { SvelteHTMLElements } from "svelte/elements";
	import MaterialIcon from "../MaterialIcon.svelte";

	let { children, ...props }: { children: Snippet } & SvelteHTMLElements["div"] = $props();
	let modalEl: HTMLElement | undefined = $state();

	$effect(() => {
		if (!modal.show || !modalEl) {
			return;
		}

		// From: https://stackoverflow.com/a/25621277
		modalEl.querySelectorAll("textarea").forEach((t) => t.dispatchEvent(new Event("input", { bubbles: true, cancelable: true })));
		const txHeight = "1rem"; // Preset initial height in pixels
		const tx = modalEl.getElementsByTagName("textarea");

		const listeners: ((e: Event) => void)[] = [];
		for (let i = 0; i < tx.length; i++) {
			if (tx[i].value === "") {
				tx[i].style.height = txHeight;
			} else {
				tx[i].style.height = tx[i].scrollHeight + "px";
			}

			tx[i].style.overflowY = "hidden";
			const listener = (e: Event) => {
				const target = e.target as HTMLTextAreaElement;
				target.style.height = "auto";
				target.style.height = target.scrollHeight + "px";
			};
			tx[i].addEventListener("input", listener, false);
			listeners.push(listener);
		}

		return () => {
			for (let i = 0; i < tx.length; i++) {
				tx[i].removeEventListener("input", listeners[i]);
			}
		};
	});
</script>

<div {...props} id="viz-modal" bind:this={modalEl}>
	<button class="modal-close-btn" onclick={() => (modal.show = false)} title="Close Modal">
		<MaterialIcon iconName="close" />
	</button>
	{@render children()}
</div>

<style lang="scss">
	#viz-modal {
		width: 45%;
		max-height: 70%;
		background-color: var(--imag-bg-color);
		z-index: 9999;
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 1em;
		position: relative;
		border-radius: 0.5em;
		box-shadow: 0 4px 20px rgba(0, 0, 0, 0.2);
	}

	.modal-close-btn {
		position: absolute;
		top: 0.75em;
		right: 0.75em;
		background: transparent;
		border: none;
		color: var(--imag-text-color);
		cursor: pointer;
		padding: 0.25em;
		border-radius: 50%;
		display: flex;
		align-items: center;
		justify-content: center;
		opacity: 0.6;
		transition:
			opacity 0.2s,
			background-color 0.2s;

		&:hover {
			opacity: 1;
			background-color: var(--imag-80);
		}
	}
</style>
