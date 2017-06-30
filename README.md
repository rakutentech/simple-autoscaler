# simple-autoscaler

Automatically scale-out/in your Cloud Foundry apps based on CPU and memory load. 

`simple-autoscaler` can be itself deployed on Cloud Foundry and is fully stateless, so it requires no database/datastore services.

This tool evolved out of dead-simple-autoscaler, a 8-lines Bash script capable of performing autoscaling for a single application:

```bash
APP_GUID=97cd2e24-81d4-4cc2-9bf2-c2ba12436afe
MIN_INST=3
MAX_INST=10
SCALE_OUT_LOAD=60 # %
SCALE_IN_LOAD=30 # %

CPU=$(cf curl /v2/apps/$APP_GUID/stats | jq '.[] | select(.state | contains("RUNNING")) | .stats.usage.cpu')
CPU_AVG=$(awk '{ s += $1 } END { print int(s/NR*100) }' <<<$CPU)
INST=$(wc -l <<<$CPU)

if   [[ $INST > $MIN_INST && $CPU_AVG < $SCALE_IN_LOAD ]]; then  
  cf curl -X PUT -d "{\"instances\":$((INST-1))}" /v2/apps/$APP_GUID?async=true
elif [[ $INST < $MAX_INST && $CPU_AVG > $SCALE_OUT_LOAD ]]; then
  cf curl -X PUT -d "{\"instances\":$((INST+1))}" /v2/apps/$APP_GUID?async=true
fi
```

Compared to this script, simple-autoscaler can do the following:

- target multiple applications
- specify application by name instead of guid
- autoscale based on cpu, memory or both

## Deploy

Deploy this application and set the following environment variables:

- `CF_API_URL`: Cloud Foundry API to target
- `CF_USERNAME`: username of the account with permissions to operate on the apps to autoscale
- `CF_PASSWORD`: password for the account above
- `AUTOSCALER_RULES`: autoscaling rules to apply (see Configuration below)

Simple autoscaler can be easily deployed on Cloud Foundry by doing the following:

- Modify the provided `manifest.yml` to set the values of the environment variables described above
- Login to Cloud Foundry and target the appropriate org/space
- Run `cf push`

## Configuration

simple-autoscaler is configured mainly through a JSON array serialized in the `AUTOSCALER_RULES` environment variable.

The JSON array should contain objects as in the following example, where we define a single CPU-based autoscaling rule for application `my_app` (space `my_space`, org `my_org`):

```json
[
  {
    "app": "my_app",
    "space": "my_space",
    "org": "my_org",
    "min_instances": 3,
    "max_instances": 10,
    "scale_in_cpu": 35,
    "scale_out_cpu": 60
  }
]
```

Each object in the array defines the autoscaling rule for one application. It is allowed to have multiple objects in the array, each referring to a different application. It is not allowed to have multiple objects refer to the _same_ application.

To set the `AUTOSCALER_RULES` variable, make sure to use valid JSON with no newlines:

```bash
cf set-env simple-autoscaler AUTOSCALER_RULES '[{"app":"my_app","space":"my_space","org": "my_org","min_instances": 3,"max_instances": 10,"scale_in_cpu": 35,"scale_out_cpu": 60}]'
```

If you have many rules it may be impractical to define the rules inline like in the example above. In this case you can store them in a separate json file and then use `jq` to inject them in your `cf` invocation:

```bash
cf set-env simple-autoscaler AUTOSCALER_RULES "$(jq -c '.' autoscaler_rules.json)"
```

Better yet, fill in your desired configuration in the provided `manifest.yml`.

### Rules

key             | description                                                      | required                               | allowed values
--------------- | ---------------------------------------------------------------- | -------------------------------------- | -----------------------------------------------------
`app`           | name of the app                                                  | required                               | existing app name
`space`         | name of the space                                                | required                               | existing space name
`org`           | name of the organization                                         | required                               | existing org name
`min_instances` | minimum number of instances the autoscaler will set              | required                               | `min_instances`>=3, `min_instances`<`max_instances`
`max_instances` | maximum number of instances the autoscaler will set              | required                               | `min_instances`<`max_instances`
`scale_in_cpu`  | average cpu load for the number of instances to be decreased     | required if `scale_out_cpu` is present | `scale_in_cpu`<`scale_out_cpu`, 0<`scale_in_cpu`<100
`scale_out_cpu` | average cpu load for the number of instances to be increased     | required if `scale_in_cpu` is present  | `scale_in_cpu`<`scale_out_cpu`, 0<`scale_out_cpu`<100
`scale_in_mem`  | average memory usage for the number of instances to be decreased | required if `scale_out_mem` is present | `scale_in_mem`<`scale_out_mem`
`scale_out_mem` | average memory usage for the number of instances to be increased | required if `scale_in_mem` is present  | `scale_in_mem`<`scale_out_mem`

- if only `scale_in_cpu` and `scale_out_cpu` are specified, autoscaling will only be based on average CPU load
- if only `scale_in_mem` and `scale_out_mem` are specified, autoscaling will only be based on average memory usage
- if all of `scale_in_cpu`, `scale_out_cpu`, `scale_in_mem` and `scale_out_mem` are specified, autoscaling will be based on both average CPU and memory usage as follows:
  - if average CPU load **or** memory usage are respectively above `scale_out_cpu`/`scale_out_mem`, the app will scale out
  - if average CPU load **and** memory usage are respectively below `scale_out_cpu`/`scale_out_mem`, the app will scale in

## Scaling policies

- The decisions to scale-out/in are based on the instantaneous average loads across all running instances.
- Scale-out/in decisions will at most increase/decrease the number of instances by 1 instance per application every 30 seconds.
- If instances for an application are crashing no decisions are made for that application.
- If the number of desired instances of an application is manually set to less than `min_instances` or to more than `max_instances`, no decisions are made for that application.

## Guidelines

Setting up autoscaling must be done carefully to avoid impacting the availability of your application.

### Common guidelines

- Make sure your app follow the 12factor app manifesto, especially factors VI, VIII and IX.
- Make sure your app, without autoscaling, does not hang or become unresponsive under load.
- Make sure application instances are ready for and start accepting traffic within 15 seconds from starting.
- simple-autoscaler makes decisions about scaling-out/in only based on CPU load, memory load and number of running instances. It does not monitor whether your application is performing correctly and it does not monitor availability/performance of external systems your application may depend on.
- Perform realistic load tests of your app when governed by the chosen autoscaler policy.
- Monitor your app performance and availability.

### CPU autoscaling

- CPU autoscaling may be unsuitable for uneven workloads (e.g. if load significantly differs between instances). Test to make sure that CPU autoscaling is appropriate for your app and workload.
- When choosing `scale_in_cpu` and `scale_out_cpu` make sure that the difference between the two is big enough to avoid flapping. A ballpark guideline is that they should be chosen as follows: `scale_in_cpu`<`scale_out_cpu`*`min_instances`/ (`min_instances`+1).
  - A good starting point is `scale_in_cpu`=`scale_out_cpu`/2.

### Memory autoscaling

- Memory autoscaling may be unsuitable for certain workloads (e.g. application using lazy GC policies that don't return memory to the OS or, more in general, applications where the memory usage does not directly depend on the actual load on the application). Test to make sure that memory autoscaling is appropriate for your app and workload.
- When choosing `scale_in_mem` and `scale_out_mem` make sure that the difference between the two is big enough to avoid flapping. A ballpark guideline is that they should be chosen as follows: `scale_in_cpu`<`scale_out_cpu`*`min_instances`/ (`min_instances`+1)
- Make sure that your application (and the runtime you are using) returns memory **to the OS** when not in use
  - Note that most GC languages don't do this: e.g. Java, Ruby and Go all don't normally return memory to the OS.

## Feedback and contributions

Feel free to open issues or PRs on Github. Before opening a PR we recommend to file a Github issue to discuss the problem you're trying to solve.

Please see the [TODO](TODO.md) doc for some possible ideas and features we would like to implement.

## Author

- Carlo Alberto Ferraris, Rakuten, Inc.

## License

[MIT](LICENSE)
