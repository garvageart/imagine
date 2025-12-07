<script lang="ts">
	import { page } from "$app/state";
	import Button from "$lib/components/Button.svelte";
	import InputText from "$lib/components/dom/InputText.svelte";
	import { registerUser } from "$lib/api";
	import { goto } from "$app/navigation";
	import { toastState } from "$lib/toast-notifcations/notif-state.svelte";

	let pageState = page.state as typeof registerData;
	let registerData = $state({
		email: "",
		password: "",
		name: ""
	});

	let notifMessage = $state("");

	function showRegNotif(
		message: string,
		level: "success" | "info" | "warning" | "error"
	) {
		notifMessage = message;
		toastState.addToast({
			message,
			type: level
		});
	}
</script>

<main
	style="background-image: url('https://picsum.photos/1920/1080/?random={Math.floor(
		Math.random() * 300
	)}');"
>
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
					showRegNotif("Please fill in all fields", "error");
					return;
				}

				if (!formObject.passwordConfirm) {
					showRegNotif("Please confirm your password", "error");
					return;
				}

				if (formObject.password !== formObject.passwordConfirm) {
					showRegNotif("Passwords do not match", "error");
					return;
				}

				try {
					const response = await registerUser({
						name: formObject.name as string,
						email: formObject.email as string,
						password: formObject.password as string
					});

					if (response.status === 201) {
						showRegNotif("Registration successful!", "success");
						goto("/auth/login");
					}
				} catch (error) {
					showRegNotif("Registration failed. Please try again.", "error");
					console.error("Registration error:", error);
				}
			}}
		>
			<InputText
				id="reg-email"
				name="email"
				label="Email"
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
			<InputText
				id="reg-password-confirm"
				name="passwordConfirm"
				placeholder="Confirm Password"
				type="password"
				required
			/>
			<Button style="margin-top: 1rem;">
				<input id="reg-submit" type="submit" value="Create" />
			</Button>
		</form>
		<p style="margin-top: 1em;">
			Already have an account? <a style="font-weight: bold;" href="/auth/login"
				>Login</a
			>
		</p>
		{#if notifMessage}
			<p style="font-size: 1.2em; font-weight: bold; margin-top: 1em;">
				{notifMessage}
			</p>
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
		color: var(--imag-100-light);
		font-family: var(--imag-code-font);
		font-weight: 700;
		font-size: 3em;
		margin: 0.5em;
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
		// max-height: 60%;
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
