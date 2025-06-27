<script>
	import { DateTime } from "luxon";
</script>

<div id="viz-test_container">
	The day today is {DateTime.now().toFormat("EEEE")}
</div>
