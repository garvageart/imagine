<script lang="ts">
	import { goto } from "$app/navigation";
	import { login } from "$lib/api";
	import Button from "$lib/components/Button.svelte";
	import InputText from "$lib/components/dom/InputText.svelte";

	let loginData = $state({
		email: "",
		password: ""
	});

	let notifMessage = $state("");

	function showLoginNotif(message: string) {
		notifMessage = message;
		setTimeout(() => (notifMessage = ""), 3000);
	}

	async function handleLogin(event: Event) {
		event.preventDefault();

		if (!loginData.email || !loginData.password) {
			showLoginNotif("Please fill in all fields");
			return;
		}

		try {
			const response = await login({
				email: loginData.email,
				password: loginData.password
			});

			if (response.status === 200) {
				showLoginNotif("Login successful!");
				goto("/");
			}
		} catch (error: any) {
			if (error.status === 401) {
				showLoginNotif("Invalid email or password");
			} else if (error.status === 404) {
				showLoginNotif("User not found");
			} else {
				showLoginNotif("Login failed. Please try again.");
			}
			console.error("Login error:", error);
		}
	}
</script>

<!-- Random background image thing needs to go at some point, maybe with a server/admin administrated background image -->
<main style="background-image: url('https://picsum.photos/1920/1080/?random={Math.floor(Math.random() * 300)}');">
	<span id="viz-title">viz</span>
	<div id="login-container">
		<h1 id="login-heading">Login</h1>
		<form id="login-form" onsubmit={handleLogin}>
			<InputText
				id="login-email"
				label="Email"
				name="email"
				placeholder="Email"
				type="email"
				required
				value={loginData.email}
				oninput={(e) => (loginData.email = e.currentTarget.value)}
			/>
			<InputText
				id="login-password"
				name="password"
				placeholder="Password"
				type="password"
				style="margin-top: 1em;"
				required
				value={loginData.password}
				oninput={(e) => (loginData.password = e.currentTarget.value)}
			/>
			<Button style="margin-top: 1rem;">
				<input id="login-submit" type="submit" value="Login" />
			</Button>
		</form>
		<p style="margin-top: 1em;">Don't have an account? <a style="font-weight: bold;" href="/auth/register">Register</a></p>
		{#if notifMessage}
			<p style="font-size: 1.2em; font-weight: bold; margin-top: 1em;">{notifMessage}</p>
		{/if}
	</div>
	<div id="login-overlay" style="height: 100%; width: 100%;"></div>
</main>

<style lang="scss">
	@use "sass:color";

	main {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		width: 100%;
		height: 100%;
		background-size: cover;
		background-position: center;
	}

	#viz-title {
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 3em;
		margin: 0.5em;
		z-index: 2;
	}

	#login-overlay {
		position: absolute;
		top: 0;
		left: 0;
		z-index: 1;
		background-color: rgb(39, 51, 74, 70%);
		backdrop-filter: blur(26px);
	}

	#login-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 3rem 2rem;
		width: 30%;
		background-color: var(--imag-bg-color);
		box-shadow: 0 -3px 0 var(--imag-primary) inset;
		z-index: 2;
	}

	#login-form {
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	:global(#login-form > input:not([type="submit"])) {
		width: 60%;
		max-width: 60%;
		min-height: 2.5rem;
		font-size: 1.5rem;
		padding: 0.5rem 2rem;
		margin-bottom: 1rem;
	}

	#login-submit {
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
