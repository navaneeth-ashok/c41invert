# c41invert

This project is a fork and merge of couple of open source projects as well as my 
own personal modifications so that the app is suitable for my way of film scanning.

## Projects referenced
- [michielbuddingh/c41invert](https://github.com/michielbuddingh/c41invert)
- [enricod/golibraw](https://github.com/enricod/golibraw)
- [inokone/golibraw](https://github.com/inokone/golibraw)

## Enhancements
The following are the enhancements over the original project by michielbuddingh  
- [x]  Option to pass an input folder instead of individual files
- [x]  Input is no longer restricted to TIFF, supports raw formats supported by libraw
- [x]  Option to pass output format, jpeg for quick conversion, TIFF for post processing.
- [ ]   Option to pass full size sampling or center weighted sampling. If you're scanning a film negative that doesn't cover the sensor area, eg 6x6 scanned on a 3:2 sensor.


## c41invert
=========

c41invert is a command-line tool to quickly convert scans of
orange-backed colour negatives into positives.  For me its main
selling point is its lack of knobs to tweak; it uses sensible defaults
to get sensible results; if a certain picture deserves a more
perfectionist approach, there is a lot of graphics software that will
help you achieve that.  This tool is meant to give you extra time
to use them.

It uses a similar technique to negfix8_, although I wasn't aware of
its existence at the time.  

### Approach

The tool samples the central section of the image, creating a
histogram of colours for each colour channel.  It then picks a
suitably 'dark' and 'light' colour (the first and ninetynineth
percentile, respectively)  

### How to use

./filmconvert convert -input ./input/ -output ./output_folder/ -output-format JPEG

The option -s-curve uses a sigmoid function (an S-shaped curve) rather
than a linear function;
