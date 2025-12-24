<script lang="ts">
	import { slide } from "svelte/transition";
	import MaterialIcon from "../MaterialIcon.svelte";
	import ChecklistFacet from "./ChecklistFacet.svelte";
	import RangeInput from "./RangeInput.svelte";
	import StarRating from "./StarRating.svelte";
	import type { ImageFilters, ImageFacets } from "$lib/states/filter.svelte";

	interface Props {
		criteria: ImageFilters;
		facets: ImageFacets;
		uiState: { expanded: Record<string, boolean> };
		save: () => void;
	}

	let {
		criteria = $bindable(),
		facets,
		uiState = $bindable(),
		save
	}: Props = $props();

	function toggleSection(section: keyof typeof uiState.expanded) {
		uiState.expanded[section] = !uiState.expanded[section];
		save();
	}
</script>

<!-- Rating -->
<div class="filter-section">
	<button class="section-header" onclick={() => toggleSection("rating")}>
		<span>Rating</span>
		<MaterialIcon
			iconName={uiState.expanded.rating
				? "keyboard_arrow_up"
				: "keyboard_arrow_down"}
			class="arrow-icon"
		/>
	</button>
	{#if uiState.expanded.rating}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			<div class="rating-row">
				<span>≥</span>
				<StarRating
					value={criteria.rating}
					onChange={(r) => {
						criteria.rating = r;
						save();
					}}
				/>
			</div>
		</div>
	{/if}
</div>

<!-- Tags -->
<div class="filter-section">
	<button class="section-header" onclick={() => toggleSection("tags")}>
		<span>Keywords</span>
		<MaterialIcon
			iconName={uiState.expanded.tags
				? "keyboard_arrow_up"
				: "keyboard_arrow_down"}
			class="arrow-icon"
		/>
	</button>
	{#if uiState.expanded.tags}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			<ChecklistFacet
				title=""
				items={facets.tags}
				selected={criteria.tags}
				onChange={(sel) => {
					criteria.tags = sel;
					save();
				}}
			/>
		</div>
	{/if}
</div>

<!-- Camera -->
<div class="filter-section">
	<button class="section-header" onclick={() => toggleSection("camera")}>
		<span>Camera</span>
		<MaterialIcon
			iconName={uiState.expanded.camera
				? "keyboard_arrow_up"
				: "keyboard_arrow_down"}
			class="arrow-icon"
		/>
	</button>
	{#if uiState.expanded.camera}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			<ChecklistFacet
				title=""
				items={facets.cameras}
				selected={criteria.camera}
				onChange={(sel) => {
					criteria.camera = sel;
					save();
				}}
			/>
		</div>
	{/if}
</div>

<!-- Lens -->
<div class="filter-section">
	<button class="section-header" onclick={() => toggleSection("lens")}>
		<span>Lens</span>
		<MaterialIcon
			iconName={uiState.expanded.lens
				? "keyboard_arrow_up"
				: "keyboard_arrow_down"}
			class="arrow-icon"
		/>
	</button>
	{#if uiState.expanded.lens}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			<ChecklistFacet
				title=""
				items={facets.lenses}
				selected={criteria.lens}
				onChange={(sel) => {
					criteria.lens = sel;
					save();
				}}
			/>
		</div>
	{/if}
</div>

<!-- Technical -->
<div class="filter-section">
	<button class="section-header" onclick={() => toggleSection("tech")}>
		<span>Technical (EXIF)</span>
		<MaterialIcon
			iconName={uiState.expanded.tech
				? "keyboard_arrow_up"
				: "keyboard_arrow_down"}
			class="arrow-icon"
		/>
	</button>
	{#if uiState.expanded.tech}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			<RangeInput
				label="ISO"
				min={facets.iso.min}
				max={facets.iso.max}
				value={criteria.iso}
				onChange={(v) => {
					criteria.iso = v;
					save();
				}}
			/>
			<RangeInput
				label="Aperture"
				min={facets.fStop.min}
				max={facets.fStop.max}
				value={criteria.fStop}
				step={0.1}
				unit="ƒ"
				onChange={(v) => {
					criteria.fStop = v;
					save();
				}}
			/>
			<RangeInput
				label="Shutter Speed"
				min={parseFloat(facets.shutterSpeed.min.toFixed(4))}
				max={parseFloat(facets.shutterSpeed.max.toFixed(4))}
				value={criteria.shutterSpeed}
				step={0.001}
				unit="s"
				onChange={(v) => {
					criteria.shutterSpeed = v;
					save();
				}}
			/>
			<RangeInput
				label="Focal Length"
				min={facets.focalLength.min}
				max={facets.focalLength.max}
				value={criteria.focalLength}
				unit="mm"
				onChange={(v) => {
					criteria.focalLength = v;
					save();
				}}
			/>
		</div>
	{/if}
</div>

<!-- Date -->
<div class="filter-section">
	<button class="section-header" onclick={() => toggleSection("date")}>
		<span>Date Taken</span>
		<MaterialIcon
			iconName={uiState.expanded.date
				? "keyboard_arrow_up"
				: "keyboard_arrow_down"}
			class="arrow-icon"
		/>
	</button>
	{#if uiState.expanded.date}
		<div class="section-content" transition:slide={{ duration: 200 }}>
			<div class="date-inputs">
				<div class="date-field">
					<span class="label">After</span>
					<input
						type="date"
						value={criteria.date.after ?? ""}
						onchange={(e) => {
							criteria.date.after =
								(e.currentTarget as HTMLInputElement).value || undefined;
							save();
						}}
					/>
				</div>
				<div class="date-field">
					<span class="label">Before</span>
					<input
						type="date"
						value={criteria.date.before ?? ""}
						onchange={(e) => {
							criteria.date.before =
								(e.currentTarget as HTMLInputElement).value || undefined;
							save();
						}}
					/>
				</div>
			</div>
		</div>
	{/if}
</div>

<style lang="scss">
	.filter-section {
		border-bottom: 1px solid var(--imag-80);

		&:last-child {
			border-bottom: none;
		}
	}

	.section-header {
		width: 100%;
		display: flex;
		justify-content: space-between;
		align-items: center;
		background: transparent;
		border: none;
		padding: 0.2rem;
		cursor: pointer;
		color: var(--imag-text-color);
		font-weight: 600;
		font-size: 0.8rem;
		text-align: left;
		transition: color 0.2s;

		&:hover {
			color: var(--imag-20);
			background-color: var(--imag-90);
		}
	}

	.section-content {
		padding: 0.5rem 0;
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.rating-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		font-size: 0.9rem;
		color: var(--imag-60);
	}

	.date-inputs {
		display: flex;
		flex-direction: column;
		gap: 0.5rem;
	}

	.date-field {
		display: flex;
		flex-direction: column;
		gap: 0.25rem;

		.label {
			font-size: 0.8rem;
			color: var(--imag-60);
		}

		input[type="date"] {
			background: var(--imag-100);
			border: none;
			box-shadow: 0 -1px 0 var(--imag-60) inset;
			color: var(--imag-text-color);
			padding: 4px 8px;
			border-radius: 0;
			font-family: var(--imag-display-font);

			&:focus {
				box-shadow: 0 -2px 0 var(--imag-primary) inset;
				outline: none;
			}

			&::-webkit-calendar-picker-indicator {
				filter: invert(1);
				opacity: 0.6;
				cursor: pointer;
			}
		}
	}

	:global(.arrow-icon) {
		font-size: 1.2rem;
		color: var(--imag-60);
	}
</style>
