<script>
	import { DateTime } from "luxon";
</script>

<div id="viz-test_container">
	The day today is <strong>{DateTime.now().toFormat("EEEE")}</strong>
</div>
