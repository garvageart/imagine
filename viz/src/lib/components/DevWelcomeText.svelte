<script lang="ts">
	import { CLIENT_IS_PRODUCTION } from "$lib/constants";
</script>

<!-- {#if CLIENT_IS_PRODUCTION} -->
<div id="welcome-text">
	<h1>Welcome to <code>viz</code></h1>
	<p>
		<code>viz</code> is the in-browser GUI for using <code>imagine</code>.
		<code>viz</code> is built using SvelteKit
	</p>
	<p>The repo for this project can be found on <strong><a href="https://github.com/garvageart/imagine">GitHub</a></strong></p>
</div>

<!-- {/if} -->

<style>
	#welcome-text {
		display: flex;
		flex-direction: column;
		justify-content: center;
		align-items: center;
		padding: 1em 0em;
	}
</style>
