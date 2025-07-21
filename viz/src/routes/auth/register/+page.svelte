<script lang="ts">
	import { goto } from "$app/navigation";
	import { login } from "$lib/states/index.svelte";

	const loginState = login.state;

	if (!loginState) {
		goto("/login");
	}
</script>

<a href="/"> Home </a>
<h1>Finish Imagine account setup</h1>
