## Comparing Cloud SQL for PostgreSQL vs Running PostgreSQL on GKE with StatefulSets + PVCs/Persistent Disks

### 1. Set up

Setting up Cloud SQL is simpler and faster as it is a dedicated service provided by Google Cloud. DIY PostgreSQL setup on the other hand requires manual implementation.

### 2. Ease of use

Integrating either one into the application is done in the same way - by supplying the connection string and credentials. However, GKE doesn't have an automatic access to the Cloud SQL. Connectivity and permission settings should be configured to access the Cloud SQL inside the cluster.

### 3. Maintainability

Cloud SQL is much easier to maintain because the underlying infrastructure is managed by Google Cloud. Replication configuration, backup tooling, restore procedures, monitoring etc. are much easier to set up in Cloud SQL.

### 4. Cost

Technically, DIY solution is cheaper as it only incurs cost for cluster resources consumption. However, managing Cloud SQL requires less time and effort. Therefore, while on the surface DIY solution might seem cheaper, when time is taken into account when calculating the cost, then Cloud SQL becomes more attractive as a solution.

### 5. When the team has multiple people

Managing permissions, so that an engineer cannot do `kubectl delete`, or alter the production Cloud SQL instance, requires configuration for both. In DIY setup, there are more layers to protect as everything is managed by the owning team. For Cloud SQL Google takes care of some of the infrastructure. Additionally, in Cloud SQL there are no K8S manifests to protect, potentially vulnerable Docker images to be aware of, or network policies to take care of. In Cloud SQL, GCP IAM roles are essentially the main thing to set up. In the case of DIY setup, individual K8S operation permissions also need to be configured, potentially taking into account team-level permissions too, so that platform team can do dangerous operations but an engineer team cannot. Additional configuration to protect the database itself are needed as well.

### 6. Scalability

Out of the box, Cloud SQL is much easier to scale horizontally and vertically. This is because Google Cloud manages the underlying infrastruture. Scaling DIY Postgres solution requires verifying that the cluster has the resources to scale. In the case of horizontal scaling, Google Cloud manages the multiple replicas itself, whereas on the DIY solution the responsibility is on the engineering team to set up and manage multiple replicas.

One advantage of DIY solution is that it is highly customizable. If a project requires some niche configuration, then it might be easier to achieve with the DIY solution .

### 7. Disaster recovery

Both solutions can achieve Disaster recovery, however, Cloud SQL is simpler to set up as it provides out of the box solution for DR. Cloud SQL provides automatic backups and Point in time Recovery solutions. In DIY solution DR can also be achieved. It can even be configured with more granular details. However, setting the solution up and testing the correctness.

### 8. Verdict

Both approaches are viable and can be setup to be production-ready. However, Cloud SQL is more of a batteries included solution in contrast to the manual configuration of DIY DB using StatefulSet with PVC. This means, Cloud SQL trades ease-of-setup for granular customizability. However, for the purposes of this project StatefulSet is a better solution for several reasons:
- There's just a single person doing the course, therefore advantages of Cloud SQL in regards to access-controls, replication, disaster recovery etc. disappear.
- Cloud SQL's additional cost is actually relevant here, since setting up a StatefulSet alongside PVC to setup a database is part of K8S learning journey.
- In this context DIY DB setup is the simpler way, eliminating the need for IAM Policy setup or cost tracking.