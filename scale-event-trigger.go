package main

/*
 * scale-event-trigger
 * (C) Copyright Bashton Ltd, 2014
 *
 * scale-event-trigger is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * scale-event-trigger is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with scale-event-trigger.  If not, see <http://www.gnu.org/licenses/>.
 *
 */
import (
	"fmt"
	"log"
	"log/syslog"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/crowdmob/goamz/aws"
	"github.com/crowdmob/goamz/ec2"
	kingpin "gopkg.in/alecthomas/kingpin.v1"
)

var (
	debug     = kingpin.Flag("debug", "Enable extra logging.").Bool()
	command   = kingpin.Flag("command", "Command to run when instances change.").Required().String()
	frequency = kingpin.Flag("frequency", "Time in seconds to wait between checking EC2 instances.").Default("60").Int()
	tags      = kingpin.Arg("tag", "Key:value pair of tags to match EC2 instances.").Strings()
	region    aws.Region
	// TaggedInstances holds the last checked set of instance IDs
	TaggedInstances = []string{}
)

func init() {
	regionname := aws.InstanceRegion()
	region = aws.Regions[regionname]
}

func main() {
	kingpin.Version("1.0.1")
	kingpin.Parse()

	sl, err := syslog.New(syslog.LOG_NOTICE|syslog.LOG_LOCAL0, "[scale-event-trigger]")
	defer sl.Close()
	if err != nil {
		log.Println("Error writing to syslog")
	} else {
		log.SetFlags(0)
		log.SetOutput(sl)
	}

	if len(*tags) == 0 {
		fmt.Println("No tags specified")
		return
	}

	// Set up access to ec2
	auth, err := aws.GetAuth("", "", "", time.Now().Add(time.Duration(24*365*time.Hour)))
	if err != nil {
		log.Println(err)
		return
	}
	ec2region := ec2.New(auth, region)

	TaggedInstances = getInstanceIDs(ec2region)
	sort.Strings(TaggedInstances)

	instanceCheck(ec2region)
}

func instanceCheck(ec2region *ec2.EC2) {
	tick := time.Tick(time.Duration(*frequency) * time.Second)
	for _ = range tick {
		instances := getInstanceIDs(ec2region)
		sort.Strings(instances)

		if *debug {
			log.Println("Stored instance list: ", TaggedInstances)
			log.Println("Fetched instance list: ", instances)
		}

		if !testEq(TaggedInstances, instances) {
			// instances are different
			log.Println("Instances changed running command")
			// run cmd
			parts := strings.Fields(*command)
			head := parts[0]
			parts = parts[1:len(parts)]

			_, err := exec.Command(head, parts...).Output()
			if err != nil {
				log.Println(err)
			} else {
				TaggedInstances = instances
			}
		}
	}
}

func testEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func getInstanceIDs(ec2region *ec2.EC2) []string {
	filter := ec2.NewFilter()

	for _, tag := range *tags {
		parts := strings.SplitN(tag, ":", 2)
		if len(parts) != 2 {
			log.Println("expected TAG:VALUE got", tag)
			break
		}
		filter.Add(fmt.Sprintf("tag:%v", parts[0]), parts[1])
	}

	taggedInstances := []string{}

	resp, err := ec2region.DescribeInstances(nil, filter)
	if err != nil {
		log.Println(err)
		return taggedInstances
	}

	for _, rsv := range resp.Reservations {
		for _, inst := range rsv.Instances {
			taggedInstances = append(taggedInstances, inst.InstanceId)
		}
	}
	return taggedInstances
}
