import os
prefix = "../raw/"
dump = "../sourceData/"
#print(sorted(os.listdir(prefix)))
arr = sorted(os.listdir(prefix))
#f = open(prefix+"/"+arr[1])

def createSnippet(signal):
    pass

def getIndexOfHighestPoint(ecg):
    index = 0
    x = ecg[0]
    i = 1
    while(i<len(ecg)):
        if(ecg[i] > x):
            index = i
            x = ecg[i]
        i+=1
    return index

def getOffset(file, begin, freq):
    i = 0
    ecg = []
    with open(prefix+file) as file_in:
        for line in file_in:
            if(i == begin + freq):
                break
            if(i >= begin):
                line = line.rstrip()
                ecg.append(float(line.split(",")[0]))
            i+=1
        return getIndexOfHighestPoint(ecg)

def filter(trigger,file,id):
  
   freq = 1000
   seconds = 10
   trig_file =  open(prefix+trigger,"r")
   arr = open(prefix+trigger,"r").readlines()
   begin = float(arr[3].rstrip())
   end = float(arr[4].rstrip())
   middle = end - begin
   offset = getOffset(file,middle,freq)
   middle += offset
   i = 0
   output = []
   with open(prefix+file) as file_in:
    for line in file_in:
        if(i == (middle + freq * seconds )):
            break
        if(i >= middle):
            output.append(line)
        i+=1
   if(file.find("_N") != -1):
       f = open(dump+"neutral/"+id,"w")
       f.writelines(output)
   elif (file.find("_H") != -1):
       f = open(dump+"happy/"+id,"w")
       f.writelines(output)
   elif (file.find("_F") != -1):
       f = open(dump+"fear/"+id,"w")
       f.writelines(output)
  
   
  
id = 0
for file in arr:
    
    if(file.find("triggers") != -1):
        filter(file,file.replace("_triggers",""),str(id))
        id += 1
    
    


