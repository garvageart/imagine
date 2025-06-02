<script lang="ts">
	import { sendOAuthParams } from "$lib/auth/auth_methods";

	const provider: string | null = new URL(location.href).searchParams.get("provider");
	sendOAuthParams(provider);
</script>

{#if !provider}
	OAuth provider not specified, redirecting back home page...
{/if}

<p>Signing in to {provider}</p>
