# Go through every Python file in the root directory and execute it as if it's a hoser
# pipeline file outputting it to a file with a .hos extension to be run by hoser.
for py in *.py; do
    [ -f "$py" ] || break
    python $py -o $py.hos
done