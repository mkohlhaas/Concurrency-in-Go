| Operation | Channel State      | Result                                                                                   |
|-----------|--------------------|------------------------------------------------------------------------------------------|
| Read      | nil                | Block                                                                                    |
|           | Open and Empty     | Block                                                                                    |
|           | Open and Not Empty | Value                                                                                    |
|           | Closed             | \<default value\>, false                                                                 |
|           | Write Only         | *Compilation Error*                                                                      |
| Write     | nil                | Block                                                                                    |
|           | Open and Full      | Block                                                                                    |
|           | Open and Not Full  | Write Value                                                                              |
|           | Closed             | **panic**                                                                                |
|           | Receive Only       | *Compilation Error*                                                                      |
| Close     | nil                | **panic**                                                                                |
|           | Open and Not Empty | Closes Channel; reads succeed until channel is drained, then reads produce default value |
|           | Open and Empty     | Closes Channel; reads produces default value                                             |
|           | Closed             | **panic**                                                                                |
|           | Receive Only       | *Compilation Error*                                                                      |
|           | Write Only         | Closes Channel                                                                           |

Closing a channel is a one-time broadcast to all receiving goroutines!
