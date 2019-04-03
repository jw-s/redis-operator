## Flow of operator

## Event types of CRD

## Added

1. Seed Master Pod is created which is a Redis instance configured as a *Master*

2. Gets the Seed Master *Pod* IP address.

4. Creates a *Service* for Sentinel pods/

5. Creates a *Endpoints* for the elected Redis master using IP address from step 2.

6. Creates a *Service* using *Endpoints* from previous step.

7. Creates a *Deployment* for Sentinels.

8. Creates a *StatefulSet* for slaves.


## Updated

1.  Deletes Seed Master Pod as no longer required as slaves now exist
2.  Queries Sentinel for current master IP Address
3.  Updates *Endpoints* with retrieved master IP
5.  Updates secondary resources with new CRD state (reconcile)


## Deleted

1. Removes tracking from operator (in memory)
3. Deletes secondary resources via API call(s)

