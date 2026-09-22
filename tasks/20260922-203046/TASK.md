# Review something pendant related to dates

- STATUS: OPEN
- PRIORITY: 30
- TAGS:

No description.

Something from Claude:

- The one gap I see: templates/new_record.html:194 collects the date from the user as dd-mm-aaaa (line 197 placeholder/pattern), but nothing in the code I've looked at yet converts that back to ISO before insert — the NewRecordHandler POST branch (handlers.go:174-178) is still empty. So when you implement the insert, you'll want the mirror-image helper: parse "02-01-2006" input → time.Parse → .Format("2006-01-02") before writing to records.date. Want me to add that (and wire up the POST handler), or are you handling that yourself?
