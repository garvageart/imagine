<script lang="ts">
	let { color }: { color?: string } = $props();
</script>

<svg
	xmlns="http://www.w3.org/2000/svg"
	viewBox="0 0 100 100"
	preserveAspectRatio="xMidYMid"
	width="64"
	height="64"
	style="shape-rendering: auto; display: block; background: transparent;"
	xmlns:xlink="http://www.w3.org/1999/xlink"
	><g
		><circle
			stroke-dasharray="164.93361431346415 56.97787143782138"
			r="35"
			stroke-width="6"
			stroke={color ? color : "var(--imag-40)"}
			fill="none"
			cy="50"
			cx="50"
		>
			<animateTransform
				keyTimes="0;1"
				values="0 50 50;360 50 50"
				dur="1s"
				repeatCount="indefinite"
				type="rotate"
				attributeName="transform"
			></animateTransform>
		</circle><g></g></g
	></svg
>
