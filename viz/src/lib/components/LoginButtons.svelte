<script lang="ts">
	import { authServerURL } from "$lib/auth/auth_methods";
	import { login } from "$lib/states/index.svelte";
	import Button from "./Button.svelte";

	const loginState = login.state;
</script>

{#if !loginState}
	<div id="login-buttons">
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
	</div>
{/if}
