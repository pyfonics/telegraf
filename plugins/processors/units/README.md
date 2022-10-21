# Units Processor Plugin

Use the `units` processor to convert the units of fields.

## Configuration

```toml @sample.conf
[[processors.units]]
    ## Regex pattern matching on field_key.
    ## hint: use regex capture groups to isolate any existing unit suffixes
    pattern = "(.*cpu_time)_ms"
    ## Unit to convert from
    from = "milliseconds"
    ## Unit to convert to
    to = "seconds"
    ## Matches of the pattern will be replaced with this string. 
    ## hint: regex capture groups can be used in the replacement string to add suffixes to.
    # replacement = "${0}" 

    ## If `unit_suffix = true` then destination unit will be added as a suffix
    # unit_suffix = true
    ## Is the field_value a monotonically increasing counter.
    # is_counter = false

    ## Tags to be added to the measurement (all values must be strings)
    # [processors.units.tags]
    #   additional_tag = "tag_value"

```

## Example

```diff

```
