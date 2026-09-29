import { Alert, AlertDescription, AlertTitle } from "@gcsim/primitives";

export function DatabaseNotice() {
	return (
		<Alert variant="warning">
			<AlertTitle>Read the configuration before you compare DPS</AlertTitle>
			<AlertDescription>
				These community simulations use different builds, targets, and
				assumptions. Results do not all follow KQM standards. Check the
				equipment and rotation before you apply a result to your team.
			</AlertDescription>
		</Alert>
	);
}
