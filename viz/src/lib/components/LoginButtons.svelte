<script lang="ts">
	import { authServerURL } from "$lib/auth/auth_methods";
	import Button from "./Button.svelte";
</script>

<div id="login-buttons">
	<Button
		id="oauth-sign_in-button-google"
		onclick={() => {
			location.href = `${authServerURL}/auth/oauth?provider=google`;
		}}
	>
		Sign in with Google
	</Button>
	<Button
		id="oauth-sign_in-button-github"
		onclick={() => {
			location.href = `${authServerURL}/auth/oauth?provider=github`;
		}}
	>
		Sign in with GitHub
	</Button>
</div>
