sqlboiler currently used is not at actual v4, and I did not bother for complete system's overhaul
so cloned from git, changed tag to v3.7.1 and built sqlboiler and sqlboiler-psql in driver directory there, after that go generate (needs the same logic, insert new schema to database before then go gen and then change code later more)

adding one role, I still used []int64, because it was quicker way to handle nil input if no role mentions were selected