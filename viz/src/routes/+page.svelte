<script lang="ts">
	import { authServerURL } from "$lib/auth/auth_methods";
	import { cookieMethods } from "../lib/utils";
	import Button from "$lib/components/Button.svelte";
	import DevWelcomeText from "$lib/components/DevWelcomeText.svelte";
	import Sidebar from "$lib/components/Sidebar.svelte";

	let loginState = cookieMethods.get("imag-state");
</script>

<h1>Welcome to <code>viz</code></h1>
<p>
	<code>viz</code> is the in-browser GUI for using <code>imagine</code>.
	<code>viz</code> is built using SvelteKit
</p>
<p>The repo for this project can be found on <a href="https://github.com/garvageart/imagine">GitHub</a></p>

{#if !loginState}
	<Button
		id="oauth-sign_in-button-google"
		onclick={() => {
			location.href = `${authServerURL}/oauth?provider=google`;
		}}
	>
		Sign in with Google
	</Button>
	<Button
		id="oauth-sign_in-button-github"
		onclick={() => {
			location.href = `${authServerURL}/oauth?provider=github`;
		}}
	>
		Sign in with GitHub
	</Button>
{/if}
