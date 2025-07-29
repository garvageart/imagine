<script lang="ts">
	import { page } from "$app/state";
	import Button from "$lib/components/Button.svelte";
	import InputText from "$lib/components/dom/InputText.svelte";
	import type { User } from "$lib/types/users";
	import { sendAPIRequest } from "$lib/utils/http";

	let pageState = page.state as typeof registerData;
	let registerData = $state({
		email: "",
		password: "",
		name: ""
	});

	let notifMessage = $state("");

	function showNotif(message: string) {
		notifMessage = message;
		setTimeout(() => (notifMessage = ""), 3000);
	}

	const randomNumber = Math.floor(Math.random() * 300);
</script>

<main style="background-image: url('https://picsum.photos/1920/1080/?random={randomNumber}');">
	<span id="viz-title">viz</span>
	<div id="reg-container">
		<h1 id="reg-heading">Register</h1>
		<form
			id="reg-form"
			onsubmit={async (event) => {
				event.preventDefault();

				// fix all this form mess. validate stuff properly lmao
				const data = new FormData(event.currentTarget);
				const formObject = Object.fromEntries(data.entries());

				if (!formObject.email || !formObject.password || !formObject.name) {
					showNotif("Please fill in all fields");
					return;
				}

				if (!formObject.passwordConfirm) {
					showNotif("Please confirm your password");
					return;
				}

				if (formObject.password !== formObject.passwordConfirm) {
					showNotif("Passwords do not match");
					return;
				}

				const response = (await sendAPIRequest<User>(
					"/user",
					{
						method: "POST",
						body: JSON.stringify(formObject),
						headers: {
							"Content-Type": "application/json"
						},
						credentials: "include",
						mode: "cors"
					},
					true
				)) as Response;
			}}
		>
			<InputText
				id="reg-email"
				label="Email"
				name="email"
				placeholder="Email"
				type="email"
				required
				disabled={pageState.email ? true : false}
				value={registerData.email}
				oninput={(e) => (registerData.email = e.currentTarget.value)}
			/>
			<InputText
				id="reg-name"
				name="name"
				placeholder="Name"
				type="text"
				required
				value={registerData.name}
				oninput={(e) => (registerData.name = e.currentTarget.value)}
			/>
			<InputText
				id="reg-password"
				name="password"
				placeholder="Password"
				type="password"
				style="margin-top: 1em;"
				required
				value={registerData.password}
				oninput={(e) => (registerData.password = e.currentTarget.value)}
			/>
			<InputText id="reg-password-confirm" name="passwordConfirm" placeholder="Password" type="password" required />
			<Button style="margin-top: 1rem;">
				<input id="reg-submit" type="submit" value="Create" />
			</Button>
		</form>
		<p style="margin-top: 1em;">Already have an account? <a style="font-weight: bold;" href="/auth/login">Login</a></p>
		{#if notifMessage}
			<p style="font-size: 1.2em; font-weight: bold; margin-top: 1em;">{notifMessage}</p>
		{/if}
	</div>
	<div id="reg-overlay" style="height: 100%; width: 100%;"></div>
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
		font-size: 4em;
		position: absolute;
		top: 1em;
		z-index: 2;
	}

	#reg-overlay {
		position: absolute;
		top: 0;
		left: 0;
		z-index: 1;
		background-color: rgb(39, 51, 74, 70%);
		backdrop-filter: blur(26px);
	}

	#reg-container {
		display: flex;
		flex-direction: column;
		align-items: center;
		padding: 3rem 2rem;
		width: 30%;
		max-height: 60%;
		background-color: var(--imag-bg-color);
		box-shadow: 0 -3px 0 var(--imag-primary) inset;
		z-index: 2;
	}

	#reg-form {
		width: 100%;
		display: flex;
		flex-direction: column;
		align-items: center;
	}

	:global(#reg-form > input:not([type="submit"])) {
		width: 60%;
		max-width: 60%;
		min-height: 2.5rem;
		font-size: 1.5rem;
		padding: 0.5rem 2rem;
		margin-bottom: 1rem;
	}

	#reg-submit {
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
