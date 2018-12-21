# My i3 status bar

Very simple status bar script. Supports these features:

- Used / Total space of root filesystem
- Used / Total system memory
- Download / Upload speed on all network interfaces
- CPU use percentage
- System load for 1, 5 and 15 minutes
- Date and time
- Colour coding filesystem, memory, CPU and load stats based on utilization

And it looks like this:

![Screenshot](https://raw.githubusercontent.com/Fornax96/i3fstatus/master/screenshot.png)

I wanted to add spotify controls too, but their D-Bus API is broken _again_.
I'll get to it when they get their shit together.

## Modifying

The code is incredibly simple and easy to understand. Anyone with basic Go
knowledge should be able to modify it. If you want more features go ahead and
fork the project. The procfs library I'm using has support for many more OS
statistics.
