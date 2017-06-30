# TODO

This document lists some ideas about features to be added or changes to be done to [simple-autoscaler](README.md).

## Linear policy

The current policy may be a little simplistic. Specifically:

- it only adds/removes one instance at a time every ~30 seconds; while removing one instance at a time is just suboptimal, adding one instance at a time may not be enough in case of sudden traffic spikes
- it may be suboptimal when it comes to scaling down: to aggressively scale down you have to decrease the difference between the scale_in/out thresholds, but doing so expooses apps to the risk of flapping

An alternative policy could be the following: we consider load as uniform between all instances and find the desired number of instances required to achieve a certain target load. As an example consider the case in which the target load is 30% and the current is 90% over 4 instances. This means that if the load was spread uniformly we would need to have 12 instances to have a load of 30%.

The following pseudo-code shows a dumb-but-correct way of doing this, while adding the ability to define minimum and maximum loads, minimum and maximum number of instances and a basic form of hysteresis to prevent flapping.

```
Icur = <current number of instances>
Lcur = <current load>
Imin, Imax = <min/max number of instances> # Imin >= 3, Imax > Imin
Lmin, Lmax = <min/max load allowed> # Lmin > Lmax
Ltgt = <target load> # Ltgt >= Lmin, Ltgt <= Lmax

Ldiffmin, Idiffmin = +inf, Imax
for I=Imin; I<=Imax; I++ {
  L = Lcur * Icur / I
  Ldiff =  Math.abs(L - Ltgt)
  if L >= Lmin && L <= Lmax && Ldiff < Ldiffmin {
    Ldiffmin = Ldiff
    Idiffmin = Idiff
  }
}
if Idiffmin == Icur - 1 {
  Idiffmin = Icur
}
return Idiffmin

# alternatively to the if above (to be able to reach Imin):
Itgtdiff = Lcur * Icur / Ltgt - Icur
if Itgtdiff > -1 && Itgtdiff < 0 {
  Idiffmin = Icur
}
```

## Traffic-based autoscaling

Autoscale based on number of requests per second. Plug into the firehose or tc and count the number of request/second.

## Service broker

Run the autoscaler as a service broker. Users would then be able to do

```bash
cf create-service autoscaler simple my_autoscaler
cf bind-service my_app my_autoscaler -c '{"scale_in_cpu":35,"scale_out_cpu":60,"min_instances":3,"max_instances":10}'
```

## Outlier instance detection

Hung instances can affect the autoscaler decisions:

- if an instance does not process requests, load can get to 0%
- if an instance spins in a loop, load can reach 100%

We should at least ignore outlier instances when computing the load average (although arguably this could be simply done by taking the median instead of the mean).

As an extension of this, we could even attempt to restart outlier instances (although arguably this would be better done by a separate reaper process)

## Scale-up/down

Add the ability to scale up/down the memory size of the container.

- Clone running my_app to stopped my_app_scaled
- Scale up/down my_app_scaled
- Start my_app_scaled

  - When all instances running:

    - Delete my_app
    - Rename my_app_scaled to my_app

  - If all instances fail to start within `timeout`:

    - Delete my_app_scaled
